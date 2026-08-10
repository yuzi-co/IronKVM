// RFC 7587 §7.1 defaults the Opus `stereo` fmtp parameter to 0. libwebrtc reads
// an absent (or zero) `stereo` as an instruction to configure a one-channel
// decoder, and it downmixes whatever arrives after that regardless of how many
// channels the sender actually encoded.
//
// The offer this browser sends is what configures its own decoder — the
// answerer cannot override that. The device's audio encoder produces genuine
// stereo (see server/service/stream/audio), so without this rewrite the
// second channel is decoded and then thrown away on every connection. This
// function edits the offer's own SDP, before it becomes the local
// description, to declare the stereo capability the browser already has.
//
// If you are reading this because you want to delete it: check first whether
// Chrome has started defaulting `stereo=1` on its own. As of this writing it
// does not, and RFC 7587 does not require it to.

/** The line-ending style found in `sdp`, used so the rewrite reproduces it exactly. */
const detectEol = (sdp: string): string => (sdp.includes('\r\n') ? '\r\n' : '\n');

const OPUS_RTPMAP = /^a=rtpmap:(\d+) opus\/48000\/2$/i;

/**
 * Rewrites an SDP offer so its Opus fmtp line declares `stereo=1;sprop-stereo=1`.
 *
 * - If an `a=fmtp` line already exists for the Opus payload type, the two
 *   parameters are appended to it.
 * - If Opus has an `a=rtpmap` line but no `a=fmtp` line, one is added.
 * - If there is no `a=rtpmap:<pt> opus/48000/2` line at all, `sdp` is
 *   returned unchanged — there is nothing to attach the parameters to.
 *
 * Every other line is left byte-for-byte alone: no reordering, no
 * reformatting, no touching the video m-section.
 */
export const withStereoOpus = (sdp: string): string => {
  const eol = detectEol(sdp);
  const lines = sdp.split(eol);

  const rtpmapIndex = lines.findIndex((line) => OPUS_RTPMAP.test(line));
  if (rtpmapIndex === -1) {
    return sdp;
  }

  const payloadType = lines[rtpmapIndex].match(OPUS_RTPMAP)?.[1];
  if (!payloadType) {
    return sdp;
  }

  const fmtpPrefix = `a=fmtp:${payloadType} `;
  const fmtpIndex = lines.findIndex((line) => line.startsWith(fmtpPrefix));

  if (fmtpIndex !== -1) {
    lines[fmtpIndex] = `${lines[fmtpIndex]};stereo=1;sprop-stereo=1`;
  } else {
    lines.splice(rtpmapIndex + 1, 0, `${fmtpPrefix}stereo=1;sprop-stereo=1`);
  }

  return lines.join(eol);
};
