import assert from 'node:assert/strict';
import { test } from 'node:test';

import { MouseReportRelative } from './mouse.ts';

// signed reads a report byte back as the int8 the host sees.
function signed(byte: number): number {
  return byte > 127 ? byte - 256 : byte;
}

function totals(reports: Uint8Array[]) {
  let x = 0;
  let y = 0;
  for (const report of reports) {
    x += signed(report[1]);
    y += signed(report[2]);
  }
  return { x, y };
}

test('a move within one report is sent as one report', () => {
  const mouse = new MouseReportRelative();
  const reports = mouse.buildMoveReports(10, -127);

  assert.equal(reports.length, 1);
  assert.deepEqual(Array.from(reports[0]), [0, 10, 129, 0, 0]);
});

test('a move past 127 is split, not clamped', () => {
  const mouse = new MouseReportRelative();
  const reports = mouse.buildMoveReports(300, 0);

  assert.equal(reports.length, 3);
  assert.deepEqual(totals(reports), { x: 300, y: 0 });
});

test('split reports stay within -127 to 127 on both axes', () => {
  const mouse = new MouseReportRelative();
  const cases: Array<[number, number]> = [
    [128, 0],
    [-128, 0],
    [0, 254],
    [255, -255],
    [-1000, 37],
    [999.6, -500.4],
    [381, 1]
  ];

  for (const [dx, dy] of cases) {
    const reports = mouse.buildMoveReports(dx, dy);
    for (const report of reports) {
      for (const byte of [report[1], report[2]]) {
        const value = signed(byte);
        assert.ok(value >= -127 && value <= 127, `${dx},${dy}: ${value}`);
      }
    }
    assert.deepEqual(totals(reports), { x: Math.round(dx), y: Math.round(dy) }, `${dx},${dy}`);
    assert.equal(
      reports.length,
      Math.ceil(Math.max(Math.abs(Math.round(dx)), Math.abs(Math.round(dy))) / 127)
    );
  }
});

test('both axes advance together, so the path keeps its direction', () => {
  const mouse = new MouseReportRelative();
  const reports = mouse.buildMoveReports(400, -200);

  for (const report of reports) {
    const x = signed(report[1]);
    const y = signed(report[2]);
    assert.ok(x > 0 && y < 0, `${x},${y}`);
    assert.ok(Math.abs(Math.abs(x) - 2 * Math.abs(y)) <= 2, `${x},${y}`);
  }
});

test('every split report carries the held buttons and no wheel', () => {
  const mouse = new MouseReportRelative();
  mouse.buttonDown(0);
  mouse.buttonDown(2);
  const reports = mouse.buildMoveReports(-500, 260);

  assert.ok(reports.length > 1);
  for (const report of reports) {
    assert.equal(report[0], 0b011);
    assert.equal(report[3], 0);
    assert.equal(report[4], 0);
  }
});

test('a move that rounds to zero still sends one report', () => {
  const mouse = new MouseReportRelative();
  const reports = mouse.buildMoveReports(0.2, -0.4);

  assert.equal(reports.length, 1);
  assert.deepEqual(Array.from(reports[0]), [0, 0, 0, 0, 0]);
});
