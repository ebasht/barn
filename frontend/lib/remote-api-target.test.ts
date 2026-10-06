import assert from "node:assert/strict";
import test from "node:test";
import {
  getRemoteApiTarget,
  isMasterLocalApiPath,
  resolveRemoteApiPath,
  setRemoteApiTarget,
} from "./remote-api-target.ts";

test("remote path rewrite proxies panel APIs only", () => {
  setRemoteApiTarget("node-1");
  assert.equal(getRemoteApiTarget(), "node-1");
  assert.equal(resolveRemoteApiPath("/api/sites"), "/api/servers/nodes/node-1/proxy/sites");
  assert.equal(
    resolveRemoteApiPath("/api/databases/x/health?full=1"),
    "/api/servers/nodes/node-1/proxy/databases/x/health?full=1",
  );
  assert.equal(resolveRemoteApiPath("/api/servers/nodes"), "/api/servers/nodes");
  assert.equal(resolveRemoteApiPath("/api/servers/settings"), "/api/servers/settings");
  assert.equal(resolveRemoteApiPath("/api/auth/qr"), "/api/auth/qr");
  assert.equal(isMasterLocalApiPath("/api/servers/overview"), true);
  setRemoteApiTarget(null);
  assert.equal(resolveRemoteApiPath("/api/sites"), "/api/sites");
});
