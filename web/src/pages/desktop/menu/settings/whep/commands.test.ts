import assert from 'node:assert/strict';
import { test } from 'node:test';

import { gstreamerCommand, obsGstreamerPipeline, shellQuote, whepEndpoint } from './commands.ts';

test('the endpoint is the page origin and the path the server reports', () => {
  assert.equal(whepEndpoint('https://kvm.local', '/api/whep'), 'https://kvm.local/api/whep');
  assert.equal(whepEndpoint('http://10.0.0.2:8080/', '/api/whep'), 'http://10.0.0.2:8080/api/whep');
  assert.equal(whepEndpoint('http://kvm', ''), 'http://kvm/api/whep');
});

test('shell quoting survives a quote in the value', () => {
  assert.equal(shellQuote('plain'), "'plain'");
  assert.equal(shellQuote("it's"), `'it'\\''s'`);
});

test('the GStreamer command plays video and audio from the endpoint with the token', () => {
  const command = gstreamerCommand('https://kvm.local/api/whep', 'nkvm_abc');
  assert.match(command, /^gst-launch-1\.0 /);
  assert.match(command, /whepclientsrc /);
  assert.match(command, /signaller::whep-endpoint='https:\/\/kvm\.local\/api\/whep'/);
  assert.match(command, /signaller::auth-token='nkvm_abc'/);
  assert.match(command, /src\.video_0 ! .*autovideosink/);
  assert.match(command, /src\.audio_0 ! .*autoaudiosink/);
});

test('the token is a placeholder until there is one', () => {
  assert.match(gstreamerCommand('http://kvm/api/whep'), /signaller::auth-token='<token>'/);
});

test('the OBS pipeline feeds the obs-gstreamer source', () => {
  const pipeline = obsGstreamerPipeline('http://kvm/api/whep', 'tok');
  assert.doesNotMatch(pipeline, /gst-launch/);
  assert.match(pipeline, /whepclientsrc /);
  assert.match(pipeline, /src\.video_0 ! .*video\.$|src\.video_0 ! .* video\. /);
  assert.match(pipeline, /src\.audio_0 ! .*audio\./);
});
