import { test, expect } from "@playwright/test";

test("home page renders core sections", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByText("WebShips", { exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: /plot your fleet/i })).toBeVisible();
  await expect(page.getByRole("button", { name: "Create Game" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Join with Code" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Leaderboard" })).toBeVisible();
});

