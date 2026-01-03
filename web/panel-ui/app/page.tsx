"use client";

import { useEffect, useState } from "react";
import { Activity, ShieldAlert, Wifi, Server, Zap, Globe, Lock } from "lucide-react";
import { StatCard } from "@/components/dashboard/StatCard";
import { EventChart } from "@/components/dashboard/EventChart";
import { RecentAlerts } from "@/components/dashboard/RecentAlerts";

// Types matching Backend API
interface DashboardStats {
    total_events: number;
    events_last_hour: number;
    alerts_count: number;
    active_agents: number;
}

export default function Dashboard() {
    const [stats, setStats] = useState<DashboardStats | null>(null);

    useEffect(() => {
        // Fetch from Go Backend
        fetch("http://localhost:8080/api/v1/dashboard/stats")
            .then((res) => res.json())
            .then((data) => setStats(data))
            .catch((err) => console.error("Failed to fetch stats:", err));
    }, []);

    // Placeholder data if backend is not reachable yet
    const displayStats = stats || {
        total_events: 125430,
        events_last_hour: 4500,
        alerts_count: 12,
        active_agents: 45
    };

    return (
        <div className="min-h-screen bg-background text-foreground p-8 font-sans selection:bg-primary/20">
            {/* Header */}
            <header className="mb-10 flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-white/10 pb-6">
                <div>
                    <h1 className="text-4xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-cyan-400">
                        S.A.K.I.N.
                    </h1>
                    <p className="text-gray-400 mt-1 flex items-center gap-2">
                        <ShieldAlert className="h-4 w-4 text-primary" />
                        Siber Analiz ve Kontrol İstihbarat Noktası
                    </p>
                </div>
                <div className="flex items-center gap-4 text-sm text-gray-400 bg-white/5 px-4 py-2 rounded-full border border-white/5">
                    <span className="flex items-center gap-2">
                        <span className="relative flex h-3 w-3">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                        </span>
                        System Online
                    </span>
                    <span className="w-px h-4 bg-gray-700"></span>
                    <span>Server: v1.0.2</span>
                </div>
            </header>

            {/* Stats Grid */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-8">
                <StatCard
                    title="Total Events"
                    value={displayStats.total_events.toLocaleString()}
                    icon={Activity}
                    description="Total processed logs"
                    trend="+12%"
                    trendUp={true}
                />
                <StatCard
                    title="Traffic (1h)"
                    value={displayStats.events_last_hour.toLocaleString()}
                    icon={Wifi}
                    description="Events per hour"
                    trend="+5%"
                    trendUp={true}
                />
                <StatCard
                    title="Critical Alerts"
                    value={displayStats.alerts_count.toString()}
                    icon={Zap}
                    description="Requires attention"
                    trend="-2%"
                    trendUp={true}
                />
                <StatCard
                    title="Active Agents"
                    value={displayStats.active_agents.toString()}
                    icon={Server}
                    description="Connected endpoints"
                    trend="Stable"
                    trendUp={true}
                />
            </div>

            {/* Main Content Grid */}
            <div className="grid gap-6 md:grid-cols-3 lg:grid-cols-3">
                <EventChart />
                <RecentAlerts />
            </div>

            {/* Footer / Status Bar placeholder */}
            <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="p-4 rounded-xl bg-card border border-white/10 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-blue-500/10 rounded-lg text-blue-400"><Globe className="h-5 w-5" /></div>
                        <div>
                            <p className="text-sm font-medium">Cyber Threat Map</p>
                            <p className="text-xs text-muted-foreground">Global intel active</p>
                        </div>
                    </div>
                    <div className="text-xs bg-blue-500/20 text-blue-300 px-2 py-1 rounded">Active</div>
                </div>
                <div className="p-4 rounded-xl bg-card border border-white/10 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-purple-500/10 rounded-lg text-purple-400"><Lock className="h-5 w-5" /></div>
                        <div>
                            <p className="text-sm font-medium">SOAR Automation</p>
                            <p className="text-xs text-muted-foreground">3 Playbooks running</p>
                        </div>
                    </div>
                    <div className="text-xs bg-purple-500/20 text-purple-300 px-2 py-1 rounded">Running</div>
                </div>
                <div className="p-4 rounded-xl bg-card border border-white/10 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-green-500/10 rounded-lg text-green-400"><Server className="h-5 w-5" /></div>
                        <div>
                            <p className="text-sm font-medium">System Health</p>
                            <p className="text-xs text-muted-foreground">All services operational</p>
                        </div>
                    </div>
                    <div className="text-xs bg-green-500/20 text-green-300 px-2 py-1 rounded">Healthy</div>
                </div>
            </div>
        </div>
    );
}
