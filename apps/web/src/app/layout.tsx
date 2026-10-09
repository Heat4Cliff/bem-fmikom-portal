import type { Metadata } from "next";
import localFont from "next/font/local";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";
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

const publicSiteOrigin = getPublicSiteOrigin();

export const metadata: Metadata = {
  metadataBase: publicSiteOrigin ? new URL(publicSiteOrigin) : undefined,
  title: {
    default: "BEM FMIKOM UNUGHA",
    template: "%s | BEM FMIKOM UNUGHA",
  },
  description: "Portal BEM Fakultas Matematika dan Ilmu Komputer UNUGHA.",
  openGraph: {
    title: "BEM FMIKOM UNUGHA",
    description: "Portal publik informasi dan kegiatan BEM FMIKOM UNUGHA.",
    type: "website",
    locale: "id_ID",
  },
  robots: { index: true, follow: true },
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id">
      <body className={`${inter.variable} ${jakarta.variable} font-sans`}>{children}</body>
    </html>
  );
}
