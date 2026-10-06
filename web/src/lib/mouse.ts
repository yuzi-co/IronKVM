// Button bit positions
const MouseButtons = {
  Left: 1 << 0,
  Right: 1 << 1,
  Middle: 1 << 2,
  Back: 1 << 3,
  Forward: 1 << 4
} as const;

// Map browser button index to HID bit
function getMouseButtonBit(button: number): number {
  switch (button) {
    case 0:
      return MouseButtons.Left;
    case 1:
      return MouseButtons.Middle;
    case 2:
      return MouseButtons.Right;
    case 3:
      return MouseButtons.Back;
    case 4:
      return MouseButtons.Forward;
    default:
      return 0;
  }
}

/**
 * Relative Mouse Report (5 bytes)
 * Used with /dev/hidg1 (relative mouse)
 *
 * Byte 0: Buttons
 * Byte 1: X movement (-127 to 127)
 * Byte 2: Y movement (-127 to 127)
 * Byte 3: Wheel (-127 to 127)
 */
export class MouseReportRelative {
  private buttons: number = 0;

  buttonDown(button: number): void {
    this.buttons |= getMouseButtonBit(button);
  }

  buttonUp(button: number): void {
    this.buttons &= ~getMouseButtonBit(button);
  }

  /**
   * Build relative mouse report
   * @param deltaX X movement (-127 to 127)
   * @param deltaY Y movement (-127 to 127)
   * @param wheel Scroll wheel (-127 to 127, negative = down)
   * @param hwheel Horizontal wheel (-127 to 127, negative = left)
   */
  buildReport(deltaX: number, deltaY: number, wheel: number = 0, hwheel = 0): Uint8Array {
    const report = new Uint8Array(5);
    report[0] = this.buttons;
    report[1] = this.clamp(Math.round(deltaX), -127, 127) & 0xff;
    report[2] = this.clamp(Math.round(deltaY), -127, 127) & 0xff;
    report[3] = this.clamp(Math.round(wheel), -127, 127) & 0xff;
    report[4] = this.clamp(Math.round(hwheel), -127, 127) & 0xff;
    return report;
  }

  /**
   * Build the reports for one pointer movement of any size.
   *
   * A report carries at most 127 counts per axis, and buildReport clamps
   * anything larger, so a fast flick used to move the host pointer less
   * than the user's mouse. The movement is split into as few reports as
   * fit, with both axes advancing together in every report so the path
   * keeps its direction. The deltas add up exactly to the rounded input.
   * Every report carries the current buttons, so a drag stays a drag.
   *
   * At least one report is returned, even for a movement that rounds to
   * zero, to match what a single buildReport call sent before.
   */
  buildMoveReports(deltaX: number, deltaY: number): Uint8Array[] {
    const x = Math.round(deltaX);
    const y = Math.round(deltaY);
    const count = Math.max(1, Math.ceil(Math.max(Math.abs(x), Math.abs(y)) / 127));

    const reports: Uint8Array[] = [];
    let sentX = 0;
    let sentY = 0;
    for (let i = 1; i <= count; i++) {
      // Cumulative targets, so rounding never drifts and the last report
      // lands exactly on the total.
      const targetX = i === count ? x : Math.round((x * i) / count);
      const targetY = i === count ? y : Math.round((y * i) / count);
      reports.push(this.buildReport(targetX - sentX, targetY - sentY));
      sentX = targetX;
      sentY = targetY;
    }
    return reports;
  }

  /**
   * Build button-only report (no movement)
   */
  buildButtonReport(): Uint8Array {
    return this.buildReport(0, 0, 0);
  }

  reset(): Uint8Array {
    this.buttons = 0;
    return this.buildReport(0, 0, 0);
  }

  private clamp(value: number, min: number, max: number): number {
    return Math.max(min, Math.min(max, value));
  }
}

/**
 * Absolute Mouse Report (6 bytes)
 * Used with /dev/hidg2 (absolute mouse/tablet)
 *
 * Byte 0: Buttons
 * Byte 1-2: X position (0 to 32767, Little Endian)
 * Byte 3-4: Y position (0 to 32767, Little Endian)
 * Byte 5: Wheel
 */
export class MouseReportAbsolute {
  private buttons: number = 0;

  buttonDown(button: number): void {
    this.buttons |= getMouseButtonBit(button);
  }

  buttonUp(button: number): void {
    this.buttons &= ~getMouseButtonBit(button);
  }

  /**
   * Build absolute mouse report
   * @param x X position (0.0 to 1.0, normalized)
   * @param y Y position (0.0 to 1.0, normalized)
   * @param wheel Scroll wheel (-127 to 127)
   */
  buildReport(x: number, y: number, wheel: number = 0): Uint8Array {
    const report = new Uint8Array(6);

    report[0] = this.buttons;
    report[1] = x & 0xff;
    report[2] = (x >> 8) & 0xff;
    report[3] = y & 0xff;
    report[4] = (y >> 8) & 0xff;
    report[5] = this.clamp(Math.round(wheel), -127, 127) & 0xff;

    return report;
  }

  /**
   * Build button-only report (keeps last position)
   */
  buildButtonReport(lastX: number, lastY: number): Uint8Array {
    return this.buildReport(lastX, lastY, 0);
  }

  reset(x: number = 0, y: number = 0): Uint8Array {
    this.buttons = 0;
    return this.buildReport(x, y, 0);
  }

  private clamp(value: number, min: number, max: number): number {
    return Math.max(min, Math.min(max, value));
  }
}

// Singleton instances
export const mouseRelative = new MouseReportRelative();
export const mouseAbsolute = new MouseReportAbsolute();
