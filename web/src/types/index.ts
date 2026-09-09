export type Resolution = {
  width: number;
  height: number;
};

export type InputRegion = {
  frameWidth: number;
  frameHeight: number;
  left: number;
  top: number;
  width: number;
  height: number;
};

// What the server holds for the capture pipeline.
//
// There is one encoder on the board, so these are device settings rather than
// browser ones: the server owns them and the menu shows what it reports. The
// browser used to keep its own copy and push it on load, which meant a second
// viewer overrode the first.
//
// quality and bitRate arrive separately because one API key carries both on the
// way in, and its value decides which of the two it means.
export type ScreenSettings = {
  width: number;
  height: number;
  quality: number;
  bitRate: number;
  fps: number;
  gop: number;
  codec: number;
};

export type ControlRegionMode = 'off' | 'auto' | 'manual';

export type OriginalResolution = Resolution;

export type ControlRegionConfig = Partial<InputRegion> & {
  mode: ControlRegionMode;
  resolutions?: OriginalResolution[];
  selectedResolution?: string;
  regions?: InputRegion[];
  selectedRegion?: string;
};
