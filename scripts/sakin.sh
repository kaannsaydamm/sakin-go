#!/bin/bash

# SGE - Master Control Script (Linux/macOS)
# Usage: ./sakin.sh {start|stop|restart|status|logs} [infra|services|all]

set -e

# Setup trap to catch Ctrl+C and just exit cleanly
trap "echo -e '\nExiting...'; exit" SIGINT

# Logs directory
mkdir -p logs

SERVICES=("ingest" "enrichment" "correlation" "analytics" "soar" "panel-api")

# ANSI Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

function show_banner() {
    clear
    echo -e "${CYAN}"
    cat << "EOF"
     _____   ___      __ __   ____  _   __       ______         ______    ___ __  _           
  ╱ ___╱  ╱   │    ╱ ╱╱_╱  ╱  _╱ ╱ │ ╱ ╱ _    ╱ ____╱___     ╱ ____╱___╱ (_) ╱_(_)___  ____ 
  ╲__ ╲  ╱ ╱│ │   ╱ ,<     ╱ ╱  ╱  │╱ ╱ (_)  ╱ ╱ __╱ __ ╲   ╱ __╱ ╱ __  ╱ ╱ __╱ ╱ __ ╲╱ __ ╲
 ___╱ ╱ ╱ ___ │_ ╱ ╱│ │_ _╱ ╱_ ╱ ╱│  ╱ _    ╱ ╱_╱ ╱ ╱_╱ ╱  ╱ ╱___╱ ╱_╱ ╱ ╱ ╱_╱ ╱ ╱_╱ ╱ ╱ ╱ ╱
╱____(_)_╱  │_(_)_╱ │_(_)___(_)_╱ │_(_│_)   ╲____╱╲____╱  ╱_____╱╲__,_╱_╱╲__╱_╱╲____╱_╱ ╱_╱ 
                                                                                            
                       
    Master Control CLI
EOF
    echo -e "${NC}"
    echo -e "${YELLOW}=========================================${NC}"
}

function load_env() {
    if [ -f .env ]; then
        set -a
        source .env
        set +a
    else
        echo -e "${YELLOW}⚠️  .env file not found! Using defaults.${NC}"
    fi
}

function start_infra() {
    echo -e "\n${CYAN}🐳 Starting Infrastructure (Docker)...${NC}"
    docker-compose up -d
    echo -e "${GREEN}✅ Infrastructure Stack is up.${NC}"
}

function stop_infra() {
    echo -e "\n${YELLOW}🛑 Stopping Infrastructure...${NC}"
    docker-compose down
}

function start_services() {
    echo -e "\n${GREEN}🚀 Starting SGE Services...${NC}"
    for svc in "${SERVICES[@]}"; do
        if pgrep -f "cmd/sge-$svc/main.go" > /dev/null; then
            echo -e "   - ${CYAN}sge-$svc${NC} is already running."
        else
            echo -e "   - Starting ${CYAN}sge-$svc${NC}..."
            nohup go run cmd/sge-$svc/main.go > logs/$svc.log 2>&1 &
        fi
    done
    echo -e "${GREEN}✅ All services started in background.${NC}"
}

function stop_services() {
    echo -e "\n${YELLOW}🛑 Stopping SGE Services...${NC}"
    pkill -f "cmd/sge-" || echo "   No services found running."
}

function logs() {
    echo -e "\n${CYAN}📄 Tailing logs (Ctrl+C to exit)...${NC}"
    tail -f logs/*.log
}

function status_check() {
    echo -e "\n${CYAN}📊 Infrastructure Status:${NC}"
    docker-compose ps
    echo -e "\n${CYAN}📊 Services Status (Running PIDs):${NC}"
    pgrep -a -f "cmd/sge-" || echo "No Go services running."
}

# --- Action Handler ---
function run_action() {
    local action=$1
    local target=$2
    
    load_env

    case "$action" in
        "start")
            if [ "$target" == "services" ]; then start_services;
            elif [ "$target" == "infra" ]; then start_infra;
            else start_infra; start_services; fi
            ;;
        "stop")
            if [ "$target" == "services" ]; then stop_services;
            elif [ "$target" == "infra" ]; then stop_infra;
            else stop_services; stop_infra; fi
            ;;
        "restart")
            stop_services
            stop_infra
            sleep 2
            start_infra
            start_services
            ;;
        "status")
            status_check
            ;;
        "logs")
            logs
            ;;
        *)
            echo "Unknown action: $action"
            ;;
    esac
}

# --- Interactive Mode vs Argument Mode ---
if [ -z "$1" ]; then
    show_banner
    echo "Select an action:"
    echo "1) Start All (Infra + Services)"
    echo "2) Stop All"
    echo "3) Restart All"
    echo "4) Status Check"
    echo "5) Tail Logs"
    echo "6) Start Services Only"
    echo "7) Stop Services Only"
    echo "8) Exit"
    
    read -p "Enter choice [1-8]: " choice
    
    case $choice in
        1) run_action "start" "all" ;;
        2) run_action "stop" "all" ;;
        3) run_action "restart" "all" ;;
        4) run_action "status" ;;
        5) run_action "logs" ;;
        6) run_action "start" "services" ;;
        7) run_action "stop" "services" ;;
        8) exit 0 ;;
        *) echo "Invalid option"; exit 1 ;;
    esac
else
    # Argument Mode
    run_action "$1" "$2"
fi
