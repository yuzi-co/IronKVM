// Plays the Opus frames that arrive on the H.264 direct websocket.
//
// The worker that reads the socket cannot play sound: AudioContext does not
// exist in a worker. So it posts each audio frame here, and this decodes it
// with WebCodecs and schedules it on a Web Audio clock.
//
// Nothing is decoded while muted. The browser will not start an AudioContext
// before the user acts, and the speaker button is that act, so the context is
// made on the first unmute and every frame before it is simply dropped.

const sampleRate = 48000;
const channels = 2;
const frameSeconds = 0.02;

// How far ahead of the audio clock a frame is scheduled. Enough to ride out
// the jitter of a websocket on a LAN, and small enough not to be heard as lag
// against the video.
const targetLatency = 0.08;

// A queue longer than this means the clock has drifted or the tab was in the
// background. Starting again at the target costs a short skip, and keeps the
// sound from trailing the picture by seconds.
const maxLatency = 0.3;

export class DirectAudioPlayer {
  private context: AudioContext | null = null;
  private decoder: AudioDecoder | null = null;
  private muted = true;
  private closed = false;
  private nextTime = 0;
  private lastSeq: number | null = null;

  static supported(): boolean {
    return typeof window.AudioDecoder === 'function' && typeof window.AudioContext === 'function';
  }

  push(seq: number, data: ArrayBuffer) {
    if (this.closed || this.muted || !this.context) {
      this.lastSeq = null;
      return;
    }

    // A sequence that goes back is a new capture on the board. A gap is frames
    // lost on the way; moving the clock across it keeps the sound in step with
    // the host rather than playing the rest of the stream early.
    if (this.lastSeq !== null) {
      if (seq <= this.lastSeq) {
        this.nextTime = 0;
      } else if (seq > this.lastSeq + 1 && this.nextTime > 0) {
        this.nextTime += (seq - this.lastSeq - 1) * frameSeconds;
      }
    }
    this.lastSeq = seq;

    const decoder = this.ensureDecoder();
    if (!decoder) {
      return;
    }

    try {
      decoder.decode(new EncodedAudioChunk({ type: 'key', timestamp: seq * 20_000, data }));
    } catch (error) {
      console.error('Direct audio decode failed:', error);
      this.resetDecoder();
    }
  }

  setMuted(muted: boolean) {
    this.muted = muted;
    if (this.closed) {
      return;
    }

    if (muted) {
      this.nextTime = 0;
      void this.context?.suspend();
      return;
    }

    if (!this.context) {
      this.context = new AudioContext({ sampleRate, latencyHint: 'interactive' });
    }
    void this.context.resume();
  }

  close() {
    this.closed = true;
    this.resetDecoder();
    void this.context?.close();
    this.context = null;
  }

  private ensureDecoder(): AudioDecoder | null {
    if (this.decoder) {
      return this.decoder;
    }

    try {
      const decoder = new AudioDecoder({
        output: (frame) => this.play(frame),
        error: (error) => {
          console.error('Direct audio decoder error:', error);
          if (this.decoder === decoder) {
            this.resetDecoder();
          }
        }
      });
      decoder.configure({ codec: 'opus', sampleRate, numberOfChannels: channels });
      this.decoder = decoder;
      return decoder;
    } catch (error) {
      console.error('Direct audio is not available in this browser:', error);
      return null;
    }
  }

  private resetDecoder() {
    if (this.decoder && this.decoder.state !== 'closed') {
      try {
        this.decoder.close();
      } catch {
        // already closed by its own error
      }
    }
    this.decoder = null;
    this.nextTime = 0;
  }

  private play(frame: AudioData) {
    const context = this.context;
    if (!context || this.muted || this.closed) {
      frame.close();
      return;
    }

    try {
      const buffer = context.createBuffer(frame.numberOfChannels, frame.numberOfFrames, frame.sampleRate);
      const plane = new Float32Array(frame.numberOfFrames);
      for (let channel = 0; channel < frame.numberOfChannels; channel++) {
        frame.copyTo(plane, { planeIndex: channel, format: 'f32-planar' });
        buffer.copyToChannel(plane, channel);
      }

      const now = context.currentTime;
      if (this.nextTime < now || this.nextTime > now + maxLatency) {
        this.nextTime = now + targetLatency;
      }

      const source = context.createBufferSource();
      source.buffer = buffer;
      source.connect(context.destination);
      source.start(this.nextTime);
      this.nextTime += buffer.duration;
    } finally {
      frame.close();
    }
  }
}
