import { defineConfig } from "@playwright/test";

/**
 * End-to-end tests: the built frontend runs in Chromium against the real Go
 * service, hosted by cmd/e2e-host (see the specification, 7.6).
 *
 * One host process holds the data, so the tests share it and run serially;
 * each test starts with a fresh planner through POST /test/reset.
 */
const port = 34115;

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  reporter: process.env.CI ? "list" : [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    viewport: { width: 1440, height: 900 },
    trace: "retain-on-failure",
  },
  webServer: {
    command: `go run ./cmd/e2e-host -addr 127.0.0.1:${port} -dist frontend/dist`,
    cwd: "..",
    url: `http://127.0.0.1:${port}/`,
    reuseExistingServer: false,
    stdout: "ignore",
    stderr: "pipe",
  },
});
