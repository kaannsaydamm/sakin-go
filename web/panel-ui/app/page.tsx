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
    const [demoMode, setDemoMode] = useState(false);

    // Fix hydration mismatch by creating specific formatter
    const formatNumber = (num: number) => num.toLocaleString('en-US');

    useEffect(() => {
        // Fetch from Go Backend
        const fetchData = async () => {
            try {
                const res = await fetch("http://localhost:8080/api/v1/dashboard/stats");
                if (res.ok) {
                    const data = await res.json();
                    setStats(data);
                }
            } catch (err) {
                console.error("Failed to fetch stats:", err);
            }
        };

        fetchData();
        const interval = setInterval(fetchData, 5000); // Refresh every 5s
        return () => clearInterval(interval);
    }, []);

    const mockStats = {
        total_events: 125430,
        events_last_hour: 4500,
        alerts_count: 12,
        active_agents: 45
    };

    const emptyStats = {
        total_events: 0,
        events_last_hour: 0,
        alerts_count: 0,
        active_agents: 0
    };

    const displayStats = demoMode ? mockStats : (stats || emptyStats);

    return (
        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-1000">
            {/* Page Title & Controls */}
            <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white mb-2">
                        Command Center
                    </h1>
                    <p className="text-gray-400 flex items-center gap-2 text-sm">
                        <span className="inline-block w-2 h-2 rounded-full bg-green-500 animate-pulse shadow-[0_0_10px_rgba(34,197,94,0.5)]"></span>
                        Real-time monitoring active
                    </p>
                </div>

                <div className="flex items-center gap-4">
                    <button
                        onClick={() => setDemoMode(!demoMode)}
                        className={`px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wider transition-all shadow-lg ${demoMode ? 'bg-blue-600 text-white shadow-blue-500/20' : 'bg-white/5 text-gray-400 hover:text-white border border-white/10'}`}
                    >
                        {demoMode ? 'Demo Mode: ON' : 'Enable Demo'}
                    </button>
                    <div className="px-4 py-2 bg-gradient-to-r from-blue-900/20 to-cyan-900/20 border border-blue-500/20 rounded-lg text-blue-400 text-sm font-mono flex items-center gap-2">
                        <Wifi className="h-4 w-4" />
                        {formatNumber(displayStats.events_last_hour)} EPS
                    </div>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                <StatCard
                    title="Total Events"
                    value={formatNumber(displayStats.total_events)}
                    icon={Activity}
                    description="Total processed logs"
                    trend="+12%"
                    trendUp={true}
                />
                <StatCard
                    title="Traffic (1h)"
                    value={formatNumber(displayStats.events_last_hour)}
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

            {/* Footer / Status Bar Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <StatusCard icon={Globe} title="Cyber Threat Map" subtitle="Global intel active" status="Active" color="blue" />
                <StatusCard icon={Lock} title="SOAR Automation" subtitle="3 Playbooks running" status="Running" color="purple" />
                <StatusCard icon={Server} title="System Health" subtitle="All services operational" status="Healthy" color="green" />
            </div>
        </div>
    );
}

function StatusCard({ icon: Icon, title, subtitle, status, color }: any) {
    const colorClasses: any = {
        blue: "text-blue-400 bg-blue-500/10 border-blue-500/20 bg-blue-500/20 text-blue-300 shadow-blue-500/10",
        purple: "text-purple-400 bg-purple-500/10 border-purple-500/20 bg-purple-500/20 text-purple-300 shadow-purple-500/10",
        green: "text-green-400 bg-green-500/10 border-green-500/20 bg-green-500/20 text-green-300 shadow-green-500/10",
    };

    // Splitting for icon bg and status badge bg
    const iconColor = colorClasses[color].split(" ").slice(0, 2).join(" ");
    const badgeColor = colorClasses[color].split(" ").slice(3).join(" ");
    const shadowColor = colorClasses[color].split(" ").pop();

    return (
        <div className={`p-4 rounded-xl bg-[#0b1221]/50 backdrop-blur-md border border-white/5 hover:border-white/10 transition-all duration-300 flex items-center justify-between group hover:shadow-lg ${shadowColor}`}>
            <div className="flex items-center gap-4">
                <div className={`p-3 rounded-xl ${iconColor} group-hover:scale-110 transition-transform duration-300`}><Icon className="h-6 w-6" /></div>
                <div>
                    <p className="text-sm font-semibold text-gray-200 group-hover:text-white transition-colors">{title}</p>
                    <p className="text-xs text-gray-500">{subtitle}</p>
                </div>
            </div>
            <div className={`text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded ${badgeColor} border border-white/5`}>{status}</div>
        </div>
    )
}
