"use client";

import { useEffect, useState } from "react";
import { Activity, ShieldAlert, Wifi, Server } from "lucide-react";

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

    if (!stats) return <div className="p-10">Loading SGE Dashboard...</div>;

    return (
        <div className="min-h-screen bg-gray-950 text-white p-8 font-sans">
            <header className="mb-8">
                <h1 className="text-3xl font-bold tracking-tight text-blue-400">S.A.K.I.N. Go Edition</h1>
                <p className="text-gray-400">Security Analytics & Knowledge Intelligence Network</p>
            </header>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <StatCard
                    title="Total Events"
                    value={stats.total_events.toLocaleString()}
                    icon={<Activity className="h-4 w-4 text-blue-500" />}
                />
                <StatCard
                    title="Events (1h)"
                    value={stats.events_last_hour.toLocaleString()}
                    icon={<Wifi className="h-4 w-4 text-green-500" />}
                />
                <StatCard
                    title="Critical Alerts"
                    value={stats.alerts_count.toString()}
                    icon={<ShieldAlert className="h-4 w-4 text-red-500" />}
                />
                <StatCard
                    title="Active Agents"
                    value={stats.active_agents.toString()}
                    icon={<Server className="h-4 w-4 text-yellow-500" />}
                />
            </div>

            <div className="mt-8 grid gap-4 md:grid-cols-2">
                <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
                    <h3 className="font-semibold mb-4">Live Traffic Preview</h3>
                    <div className="text-sm text-gray-500">Charts would be rendered here using Recharts...</div>
                </div>
                <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
                    <h3 className="font-semibold mb-4">Recent Alerts</h3>
                    <div className="text-sm text-gray-500">Alert list would go here...</div>
                </div>
            </div>
        </div>
    );
}

function StatCard({ title, value, icon }: { title: string; value: string; icon: any }) {
    return (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-6 shadow-sm">
            <div className="flex flex-row items-center justify-between space-y-0 pb-2">
                <span className="text-sm font-medium text-gray-200">{title}</span>
                {icon}
            </div>
            <div className="text-2xl font-bold">{value}</div>
        </div>
    );
}
