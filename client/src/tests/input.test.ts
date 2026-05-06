import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { InputHandler } from "../input";

describe("InputHandler", () => {
  let handler: InputHandler;

  beforeEach(() => { handler = new InputHandler(); });
  afterEach(() => { handler.stop(); });

  it("starts with all keys false", () => {
    const inp = handler.getInput(0);
    expect(inp.keys.up).toBe(false);
    expect(inp.keys.down).toBe(false);
    expect(inp.keys.left).toBe(false);
    expect(inp.keys.right).toBe(false);
    expect(inp.keys.space).toBe(false);
  });

  it("tracks keydown/keyup for WASD", () => {
    handler.simulateKeyDown("w");
    expect(handler.getInput(1).keys.up).toBe(true);

    handler.simulateKeyUp("w");
    expect(handler.getInput(2).keys.up).toBe(false);
  });

  it("tracks arrow keys", () => {
    handler.simulateKeyDown("ArrowLeft");
    expect(handler.getInput(1).keys.left).toBe(true);
  });

  it("tracks space", () => {
    handler.simulateKeyDown(" ");
    expect(handler.getInput(1).keys.space).toBe(true);
  });

  it("captures mouse clicks (wired but no game binding)", () => {
    handler.simulateMouseClick(100, 200);
    const inp = handler.getInput(1);
    expect(inp.mouse).toBeDefined();
    expect(inp.mouse!.x).toBe(100);
    expect(inp.mouse!.y).toBe(200);
    expect(inp.mouse!.click).toBe(true);
  });

  it("getInput sets the tick field", () => {
    expect(handler.getInput(7).tick).toBe(7);
  });
});
