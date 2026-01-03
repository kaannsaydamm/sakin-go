"use client";

import { Bell, Search, Settings } from "lucide-react";

export function Header() {
    return (
        <header className="h-20 border-b border-white/5 bg-[#030919]/80 backdrop-blur-xl sticky top-0 z-40 px-8 flex items-center justify-between">
            {/* Search Bar */}
            <div className="flex-1 max-w-xl">
                <div className="relative group">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500 group-focus-within:text-blue-400 transition-colors" />
                    <input
                        type="text"
                        placeholder="Search IP, Hash, Event ID..."
                        className="w-full h-10 pl-11 pr-4 rounded-full bg-white/5 border border-white/5 text-sm text-white placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:bg-white/10 transition-all shadow-inner"
                    />
                </div>
            </div>

            {/* Right Actions */}
            <div className="flex items-center gap-6">
                {/* System Ticker */}
                <div className="hidden lg:flex items-center gap-2 px-4 py-1.5 rounded-full bg-blue-900/20 border border-blue-500/20">
                    <span className="relative flex h-2 w-2">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                        <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                    </span>
                    <span className="text-xs font-mono text-blue-300 tracking-wider">SYSTEM OPTIMAL</span>
                </div>

                <button className="relative p-2 rounded-full hover:bg-white/5 transition-colors text-gray-400 hover:text-white">
                    <Bell className="h-5 w-5" />
                    <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-red-500 border border-black shadow-[0_0_10px_#ef4444]" />
                </button>

                <button className="p-2 rounded-full hover:bg-white/5 transition-colors text-gray-400 hover:text-white">
                    <Settings className="h-5 w-5" />
                </button>
            </div>
        </header>
    );
}
