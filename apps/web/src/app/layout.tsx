import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";

const inter = localFont({
  src: "../../node_modules/@fontsource-variable/inter/files/inter-latin-wght-normal.woff2",
  variable: "--font-inter",
  display: "swap",
  weight: "100 900",
});
const jakarta = localFont({
  src: "../../node_modules/@fontsource-variable/plus-jakarta-sans/files/plus-jakarta-sans-latin-wght-normal.woff2",
  variable: "--font-jakarta",
  display: "swap",
  weight: "200 800",
});

export const metadata: Metadata = {
  title: {
    default: "BEM FMIKOM UNUGHA",
    template: "%s | BEM FMIKOM UNUGHA",
  },
  description: "Portal BEM Fakultas Matematika dan Ilmu Komputer UNUGHA.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id">
      <body className={`${inter.variable} ${jakarta.variable} font-sans`}>{children}</body>
    </html>
  );
}
