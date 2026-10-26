// ICE candidate ordering for the H.264 WebRTC session (ironkvm-dist#155).
//
// This browser offers and the board answers. Both directions trickle their
// candidates over the one signalling socket, and each side can see the other's
// candidates before the description they belong to:
//
// - This browser can surface a candidate before its offer goes out. The spec
//   orders the setLocalDescription result first, but a pooled candidate or a
//   browser that fires icecandidate early does not wait. The board's pion
//   refuses a candidate before the offer, so CandidateOutbox holds them until
//   the offer is sent. The server holds them as well, for older clients.
// - The board starts gathering when it sets its answer and sends candidates
//   from another goroutine, so one can arrive here before the answer.
//   CandidateInbox holds those until the answer is applied.

/** Holds this browser's candidates until its offer is sent, then sends each at once. */
export class CandidateOutbox {
  private held: RTCIceCandidateInit[] = [];
  private open = false;
  private gathered = 0;
  private readonly send: (candidate: RTCIceCandidateInit) => boolean;

  constructor(send: (candidate: RTCIceCandidateInit) => boolean) {
    this.send = send;
  }

  /** How many candidates this browser gathered for the session. */
  get count(): number {
    return this.gathered;
  }

  /**
   * Takes a gathered candidate. False when sending it failed. Firefox ends
   * gathering with an empty candidate, which goes on but is not counted.
   */
  add(candidate: RTCIceCandidateInit): boolean {
    if (candidate.candidate) {
      this.gathered += 1;
    }
    if (!this.open) {
      this.held.push(candidate);
      return true;
    }
    return this.send(candidate);
  }

  /** The offer went out: sends the held candidates. False when one failed. */
  offerSent(): boolean {
    this.open = true;
    const held = this.held;
    this.held = [];
    return held.every((candidate) => this.send(candidate));
  }
}

/** Holds the board's candidates until its answer is applied. */
export class CandidateInbox<T> {
  private held: T[] = [];
  private open = false;
  private readonly add: (candidate: T) => void;

  constructor(add: (candidate: T) => void) {
    this.add = add;
  }

  /**
   * Takes a candidate from the board. hasRemoteDescription is the peer's own
   * view, which can be set a moment before answerApplied runs.
   */
  receive(candidate: T, hasRemoteDescription: boolean): void {
    if (this.open || hasRemoteDescription) {
      this.add(candidate);
      return;
    }
    this.held.push(candidate);
  }

  /** The answer is set: adds the held candidates, in order. */
  answerApplied(): void {
    this.open = true;
    const held = this.held;
    this.held = [];
    held.forEach((candidate) => this.add(candidate));
  }
}

/**
 * Why a session failed, when it is a reason the user can act on:
 * - no-candidates: the browser gathered no address at all to offer, which a
 *   browser policy or extension limiting WebRTC causes, or a network with no
 *   usable interface. A STUN server can give a public address to offer.
 * - no-route: both sides exchanged addresses but no check got through, as from
 *   another network without STUN, a VPN or a forwarded port.
 * Anything else is null and gets the general message.
 */
export type IceFailure = 'no-candidates' | 'no-route' | null;

export type IceFailureState = {
  gatheringState: RTCIceGatheringState;
  localCandidates: number;
  answerApplied: boolean;
  connectionState: RTCPeerConnectionState;
};

export const iceFailure = (state: IceFailureState): IceFailure => {
  if (state.gatheringState === 'complete' && state.localCandidates === 0) {
    return 'no-candidates';
  }
  if (!state.answerApplied) {
    return null;
  }
  switch (state.connectionState) {
    case 'new':
    case 'connecting':
    case 'failed':
      return 'no-route';
    default:
      return null;
  }
};
