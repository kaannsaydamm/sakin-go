"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ShieldAlert, AlertTriangle, AlertCircle } from "lucide-react";

const alerts = [
    { id: 1, title: "Brute Force Detected", source: "192.168.1.50", severity: "Critical", time: "2 min ago" },
    { id: 2, title: "Malicious IP Connection", source: "10.0.0.5", severity: "High", time: "15 min ago" },
    { id: 3, title: "Port Scan Activity", source: "192.168.1.102", severity: "Medium", time: "1 hour ago" },
    { id: 4, title: "New Admin User Created", source: "Internal", severity: "Low", time: "3 hours ago" },
];

export function RecentAlerts() {
    return (
        <Card className="col-span-1 bg-card/50 backdrop-blur-sm border-white/10">
            <CardHeader>
                <CardTitle>Recent Critical Alerts</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    {alerts.map((alert) => (
                        <div key={alert.id} className="flex items-center justify-between p-3 rounded-lg bg-white/5 hover:bg-white/10 transition-colors border border-white/5">
                            <div className="flex items-center space-x-3">
                                <SeverityIcon severity={alert.severity} />
                                <div>
                                    <p className="text-sm font-medium text-white">{alert.title}</p>
                                    <p className="text-xs text-muted-foreground">{alert.source}</p>
                                </div>
                            </div>
                            <span className="text-xs text-muted-foreground">{alert.time}</span>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
}

function SeverityIcon({ severity }: { severity: string }) {
    switch (severity) {
        case "Critical":
            return <ShieldAlert className="h-5 w-5 text-red-500" />;
        case "High":
            return <AlertTriangle className="h-5 w-5 text-orange-500" />;
        case "Medium":
            return <AlertCircle className="h-5 w-5 text-yellow-500" />;
        default:
            return <ShieldAlert className="h-5 w-5 text-blue-500" />;
    }
}
