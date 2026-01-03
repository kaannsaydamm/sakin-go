"use client";

import { Activity, Globe, Lock, Server, ShieldAlert, Zap } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const navItems = [
    { title: "Dashboard", href: "/", icon: Activity },
    { title: "Traffic Explorer", href: "/traffic", icon: Globe },
    { title: "Agent Manager", href: "/agents", icon: Server },
    { title: "SOAR Platform", href: "/soar", icon: Zap },
    { title: "SIEM Rules", href: "/rules", icon: Lock },
];

export function Sidebar() {
    const pathname = usePathname();

    return (
        <aside className="w-72 border-r border-white/5 bg-[#030919]/80 backdrop-blur-2xl fixed inset-y-0 left-0 z-50 flex flex-col shadow-2xl shadow-blue-900/5">
            {/* Logo */}
            <div className="h-20 flex items-center px-8 border-b border-white/5 bg-gradient-to-r from-blue-900/10 to-transparent">
                <div className="flex items-center gap-3 font-bold text-2xl tracking-widest text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-cyan-300">
                    <ShieldAlert className="h-8 w-8 text-blue-500 drop-shadow-[0_0_15px_rgba(59,130,246,0.5)]" />
                    S.A.K.I.N.
                </div>
            </div>

            {/* Nav */}
            <nav className="flex-1 px-4 py-8 space-y-2 overflow-y-auto">
                {navItems.map((item) => {
                    const isActive = pathname === item.href;
                    return (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={cn(
                                "flex items-center px-4 py-4 rounded-xl transition-all duration-300 group relative overflow-hidden",
                                isActive
                                    ? "bg-blue-600/10 text-blue-300 shadow-[0_0_20px_rgba(37,99,235,0.1)] border border-blue-500/20"
                                    : "text-gray-400 hover:text-white hover:bg-white/5"
                            )}
                        >
                            {isActive && (
                                <div className="absolute left-0 top-0 bottom-0 w-1 bg-blue-500 shadow-[0_0_10px_#3b82f6] rounded-r-full" />
                            )}
                            <item.icon className={cn(
                                "mr-4 h-5 w-5 transition-transform duration-300 group-hover:scale-110",
                                isActive ? "text-blue-400 drop-shadow-[0_0_8px_rgba(96,165,250,0.5)]" : "group-hover:text-blue-400"
                            )} />
                            <span className="font-medium tracking-wide">{item.title}</span>

                            {/* Hover Glow */}
                            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/0 via-blue-500/5 to-blue-500/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000" />
                        </Link>
                    );
                })}
            </nav>

            {/* Footer User Profile */}
            <div className="p-6 border-t border-white/5 bg-black/20">
                <div className="flex items-center gap-4 p-3 rounded-2xl bg-white/5 border border-white/5 hover:border-blue-500/30 transition-colors cursor-pointer group">
                    <div className="h-10 w-10 rounded-full bg-gradient-to-tr from-blue-600 to-cyan-500 p-[2px] shadow-lg shadow-blue-500/20 group-hover:shadow-blue-500/40 transition-shadow">
                        <div className="h-full w-full rounded-full bg-[#0a0a0a] flex items-center justify-center">
                            <span className="font-bold text-xs text-blue-200">KS</span>
                        </div>
                    </div>
                    <div>
                        <p className="text-sm font-semibold text-white group-hover:text-blue-300 transition-colors">Kaan S.</p>
                        <p className="text-xs text-blue-400/70">SOC Commander</p>
                    </div>
                </div>
            </div>
        </aside>
    );
}
