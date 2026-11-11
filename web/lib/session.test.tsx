import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { notifySessionChanged, readSession, saveSession, useSession } from "./session";

const user = { token: "t", userId: "3", role: "Editor", name: "An", email: "an@cms.com" };

beforeEach(() => localStorage.clear());

describe("session", () => {
  it("is signed out when nobody is signed in", () => {
    expect(readSession()).toMatchObject({ signedIn: false, token: "", role: "author" });
    expect(renderHook(() => useSession()).result.current.signedIn).toBe(false);
  });

  it("keeps the stored role even without a token", () => {
    localStorage.setItem("userRole", "Admin");
    expect(readSession()).toMatchObject({ signedIn: false, role: "admin" });
  });

  it("saves a session and lowercases the role", () => {
    saveSession(user);
    expect(readSession()).toEqual({ ...user, signedIn: true, role: "editor" });
  });

  it("defaults role and email when they are missing", () => {
    localStorage.setItem("token", "t");
    expect(readSession()).toMatchObject({ signedIn: true, role: "author", email: "user@cms.com" });
  });

  it("updates mounted hooks when the session is saved or changed elsewhere", () => {
    const { result } = renderHook(() => useSession());
    act(() => saveSession(user));
    expect(result.current.userId).toBe("3");
    act(() => {
      localStorage.clear();
      notifySessionChanged();
    });
    expect(result.current.signedIn).toBe(false);
  });
});
