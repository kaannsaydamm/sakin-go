import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import { Sidebar } from "@/components/layout/Sidebar";
import { Header } from "@/components/layout/Header";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
});

export const metadata: Metadata = {
  title: "S.A.K.I.N. | Security Operations Center",
  description: "Advanced Cyber Security Dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${jetbrainsMono.variable} font-sans bg-[#020817] text-white overflow-hidden`}>
        {/* Background Effects */}
        <div className="fixed inset-0 z-[-1] bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-blue-900/20 via-[#020817] to-[#020817]" />
        <div className="fixed inset-0 z-[-1] bg-[url('/noise.png')] opacity-[0.03] mix-blend-overlay pointer-events-none" />

        <div className="flex h-screen">
          <Sidebar />
          <div className="flex-1 flex flex-col pl-72">
            <Header />
            <main className="flex-1 overflow-y-auto p-8 relative">
              {/* Main Content Glow */}
              <div className="absolute top-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500/20 to-transparent" />
              {children}
            </main>
          </div>
        </div>
      </body>
    </html>
  );
}
