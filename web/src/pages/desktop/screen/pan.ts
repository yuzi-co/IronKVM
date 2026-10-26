// Panning a zoomed picture. The Scale menu enlarges the viewport about its
// centre, and the container clips it, so at 150 % or 200 % only the middle
// of the host screen shows. A pan moves the enlarged viewport inside the
// container to bring the rest into view.
//
// A pan is kept as a fraction of the travel on each axis, from -1 to 1, with
// 0 centred. It then survives a resize or a change of scale without pushing
// the picture past its edge: the travel is worked out again and the fraction
// still lands inside it.

export type Size = {
  width: number;
  height: number;
};

export type Pan = {
  x: number;
  y: number;
};

export const centred: Pan = { x: 0, y: 0 };

// fitContain is the size of a picture of size media drawn inside box with
// object-fit: contain. Without a usable media size the picture fills the box.
export function fitContain(box: Size, media: Size | null | undefined): Size {
  if (!media || !(media.width > 0) || !(media.height > 0)) return box;
  if (!(box.width > 0) || !(box.height > 0)) return box;
  const ratio = Math.min(box.width / media.width, box.height / media.height);
  return { width: media.width * ratio, height: media.height * ratio };
}

// panTravel is how far, in CSS pixels, the enlarged picture can move each
// way from the centre before its edge comes into the container. The picture
// is the part of the viewport that shows the host, without the letterbox
// bars, and is centred in it. An axis on which it fits has no travel.
export function panTravel(picture: Size, container: Size, scale: number): Pan {
  if (!(scale > 0)) return { x: 0, y: 0 };
  return {
    x: Math.max(0, (picture.width * scale - container.width) / 2),
    y: Math.max(0, (picture.height * scale - container.height) / 2)
  };
}

function clampUnit(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.max(-1, Math.min(1, value));
}

export function clampPan(pan: Pan): Pan {
  return { x: clampUnit(pan.x), y: clampUnit(pan.y) };
}

// dragPan moves the pan by a drag of dx, dy CSS pixels. The picture follows
// the finger, as if it were grabbed: a drag to the right shows more of the
// left side. An axis without travel stays centred.
export function dragPan(pan: Pan, dx: number, dy: number, travel: Pan): Pan {
  return clampPan({
    x: travel.x > 0 ? pan.x + dx / travel.x : 0,
    y: travel.y > 0 ? pan.y + dy / travel.y : 0
  });
}

// panOffset is the translation, in CSS pixels, that puts the pan on screen.
export function panOffset(pan: Pan, travel: Pan): Pan {
  const clamped = clampPan(pan);
  // Adding 0 turns a -0 into 0, which keeps the transform string tidy.
  return { x: clamped.x * travel.x + 0, y: clamped.y * travel.y + 0 };
}

export function isPannable(travel: Pan): boolean {
  return travel.x > 0.5 || travel.y > 0.5;
}
