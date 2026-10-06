import assert from "node:assert/strict";
import test from "node:test";
import {
  GLOBAL_BARN_SCOPE,
  normalizeBarnTargets,
  parseBarnScope,
  serializeBarnScope,
  withBarnScope,
} from "./barn-scope.ts";
import { resolveBarnNavigation } from "./barn-navigation.ts";
import type { ServerNode, ServersSettings } from "./types.ts";

const settings: ServersSettings = {
  mode: "master",
  node_uid: "master-uid",
  node_name: "Main Barn",
  public_url: "",
  master_url: "",
  notification_mode: "local",
  has_master_token: false,
};

const localNode: ServerNode = {
  id: "local-id",
  node_uid: "master-uid",
  name: "old name",
  role: "master",
  connection_type: "local",
  base_url: "",
  status: "online",
  version: "dev",
  capabilities: [],
  open_incidents: 0,
};

test("scope parses, serializes and updates URLs", () => {
  assert.deepEqual(parseBarnScope(null), GLOBAL_BARN_SCOPE);
  assert.deepEqual(parseBarnScope("node-1"), { type: "node", nodeId: "node-1" });
  assert.equal(serializeBarnScope({ type: "node", nodeId: "node-1" }), "node-1");
  assert.equal(withBarnScope("/sites?view=grid", { type: "node", nodeId: "node-1" }), "/sites?view=grid&barn=node-1");
  assert.deepEqual(parseBarnScope("all"), GLOBAL_BARN_SCOPE);
  assert.equal(withBarnScope("/sites?barn=old", GLOBAL_BARN_SCOPE), "/sites?barn=all");
});

test("local target has one canonical adapter", () => {
  assert.deepEqual(normalizeBarnTargets(settings, [localNode])[0], {
    id: "local-id",
    name: "Main Barn",
    kind: "local",
    status: "online",
    isMaster: true,
    baseUrl: undefined,
    capabilities: [],
  });
});

test("navigation follows mode and target capability", () => {
  assert.equal(resolveBarnNavigation(true, null)[0]?.href, "/overview");
  const local = normalizeBarnTargets(settings, [localNode])[0];
  assert.equal(resolveBarnNavigation(true, local).some((item) => item.href === "/databases"), true);
  assert.equal(
    resolveBarnNavigation(true, { ...local, id: "remote", kind: "agent", isMaster: false }).some((item) => item.href === "/databases"),
    false,
  );
});
