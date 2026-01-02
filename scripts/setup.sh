#!/bin/bash
# SGE - Sakin Go Edition: Interactive Setup Script
# Usage: ./setup.sh

set -e

# ANSI Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function show_banner() {
    clear
    echo -e "${BLUE}"
    cat << "EOF"
   _____          _  __ _____ _   _     _____  ____  
     _____   ___      __ __   ____  _   __       ______         ______    ___ __  _           
  ╱ ___╱  ╱   │    ╱ ╱╱_╱  ╱  _╱ ╱ │ ╱ ╱ _    ╱ ____╱___     ╱ ____╱___╱ (_) ╱_(_)___  ____ 
  ╲__ ╲  ╱ ╱│ │   ╱ ,<     ╱ ╱  ╱  │╱ ╱ (_)  ╱ ╱ __╱ __ ╲   ╱ __╱ ╱ __  ╱ ╱ __╱ ╱ __ ╲╱ __ ╲
 ___╱ ╱ ╱ ___ │_ ╱ ╱│ │_ _╱ ╱_ ╱ ╱│  ╱ _    ╱ ╱_╱ ╱ ╱_╱ ╱  ╱ ╱___╱ ╱_╱ ╱ ╱ ╱_╱ ╱ ╱_╱ ╱ ╱ ╱ ╱
╱____(_)_╱  │_(_)_╱ │_(_)___(_)_╱ │_(_│_)   ╲____╱╲____╱  ╱_____╱╲__,_╱_╱╲__╱_╱╲____╱_╱ ╱_╱ 
                                                                                            
                                                     
      Sakin: Go Edition - Infrastructure Setup
EOF
    echo -e "${NC}"
    echo -e "${YELLOW}=====================================================${NC}"
}

function check_dependencies() {
    echo -e "\n${BLUE}[1/3] Checking Dependencies...${NC}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed!${NC} Please install Go 1.22+"
        exit 1
    else
        GO_VERSION=$(go version)
        echo -e "${GREEN}✅ Go is installed:${NC} $GO_VERSION"
    fi

    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed!${NC} Please install Docker Desktop or Engine."
        exit 1
    else
        echo -e "${GREEN}✅ Docker is installed.${NC}"
    fi

    if ! command -v docker-compose &> /dev/null; then
        echo -e "${YELLOW}⚠️  docker-compose not found (might be a plugin 'docker compose'). Checking plugin...${NC}"
        if docker compose version &> /dev/null; then
             echo -e "${GREEN}✅ Docker Compose plugin found.${NC}"
        else
             echo -e "${RED}❌ Docker Compose not found.${NC}"
             exit 1
        fi
    else
        echo -e "${GREEN}✅ Docker Compose is installed.${NC}"
    fi
}

function install_go_modules() {
    echo -e "\n${BLUE}[2/3] Installing Go Modules...${NC}"
    go mod tidy
    go mod download
    echo -e "${GREEN}✅ Modules downloaded.${NC}"
}

function setup_certs() {
    echo -e "\n${BLUE}[3/3] Setting up Certificates & Directories...${NC}"
    
    mkdir -p certs logs
    
    if [ ! -f .env ]; then
        echo -e "${YELLOW}⚠️  .env file missing. Creating from example...${NC}"
        if [ -f .env.example ]; then
            cp .env.example .env
            echo -e "${GREEN}✅ Created .env from .env.example.${NC}"
        fi
    fi

    echo -e "${GREEN}✅ Directory structure ready.${NC}"
}

function full_setup() {
    check_dependencies
    install_go_modules
    setup_certs
    echo -e "\n${GREEN}🎉 Setup Complete! You can now run './scripts/sakin.sh' to start.${NC}"
}

# Interactive Menu
show_banner
echo "Select an action:"
echo "1) Full Setup (Dependencies + Modules + Certs)"
echo "2) Install Go Modules Only"
echo "3) Setup Directories & .env Only"
echo "4) Exit"
read -p "Enter choice [1-4]: " choice

case $choice in
    1)
        full_setup
        ;;
    2)
        install_go_modules
        ;;
    3)
        setup_certs
        ;;
    4)
        echo "Exiting."
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid option.${NC}"
        exit 1
        ;;
esac
