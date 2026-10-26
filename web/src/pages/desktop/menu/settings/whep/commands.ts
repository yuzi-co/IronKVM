// The WHEP endpoint's URL and the commands that play it (ironkvm-dist#145).
// Pure, so node --test can check them.

const defaultPath = '/api/whep';

// The placeholder shown where the player's token goes.
export const TOKEN_PLACEHOLDER = '<token>';

// whepEndpoint is where a player POSTs its offer: this page's origin and the
// path the server reports.
export function whepEndpoint(origin: string, path: string): string {
  return `${origin.replace(/\/+$/, '')}${path || defaultPath}`;
}

// shellQuote quotes a value for a POSIX shell.
export function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

// The source element and its links, shared by the command and the OBS
// pipeline. whepclientsrc (gst-plugins-rs) negotiates and decodes; its pads
// appear as the tracks arrive, video_0 always and audio_0 when the board
// sends the host's audio.
function source(endpoint: string, token: string) {
  return `whepclientsrc signaller::whep-endpoint=${shellQuote(endpoint)} signaller::auth-token=${shellQuote(token)} name=src`;
}

// gstreamerCommand plays the board's video and audio in a window.
export function gstreamerCommand(endpoint: string, token: string = TOKEN_PLACEHOLDER): string {
  return [
    `gst-launch-1.0 ${source(endpoint, token)}`,
    'src.video_0 ! queue ! videoconvert ! autovideosink',
    'src.audio_0 ! queue ! audioconvert ! audioresample ! autoaudiosink'
  ].join(' \\\n  ');
}

// obsGstreamerPipeline is the pipeline for OBS's GStreamer Source
// (obs-gstreamer plugin), which takes the elements named video and audio as
// its sinks. OBS has no WHEP source of its own.
export function obsGstreamerPipeline(endpoint: string, token: string = TOKEN_PLACEHOLDER): string {
  return [
    source(endpoint, token),
    'src.video_0 ! queue ! videoconvert ! video.',
    'src.audio_0 ! queue ! audioconvert ! audioresample ! audio.'
  ].join(' ');
}
