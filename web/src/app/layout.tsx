import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "WebShips",
  description: "Battleship duels with real-time play.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
