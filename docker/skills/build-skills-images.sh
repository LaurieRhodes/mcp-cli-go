#!/bin/bash
# Build script for skills container images
# Usage: ./build-skills-images.sh [image-name]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Skills Container Image Builder${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Detect container runtime (prefer podman for rootless)
if command -v podman &> /dev/null; then
    RUNTIME="podman"
    echo -e "${GREEN}✓${NC} Using Podman"
elif command -v docker &> /dev/null; then
    RUNTIME="docker"
    echo -e "${GREEN}✓${NC} Using Docker"
else
    echo -e "${RED}✗${NC} No container runtime found (need docker or podman)"
    exit 1
fi

# Image definitions
declare -A IMAGES
IMAGES[docx]="Dockerfile.docx:mcp-skills-docx:DOCX skill (OOXML manipulation)"
IMAGES[pptx]="Dockerfile.pptx:mcp-skills-pptx:PPTX skill (PowerPoint)"
IMAGES[xlsx]="Dockerfile.xlsx:mcp-skills-xlsx:XLSX skill (basic Excel)"
IMAGES[xlsx-libreoffice]="Dockerfile.xlsx-libreoffice:mcp-skills-xlsx-libreoffice:XLSX + LibreOffice (formulas)"
IMAGES[pdf]="Dockerfile.pdf:mcp-skills-pdf:PDF skill (PDF manipulation)"
IMAGES[office]="Dockerfile.office:mcp-skills-office:Combined (all Office formats)"
IMAGES[document-parsing]="Dockerfile.document-parsing:mcp-skills-document-parsing:Document parsing (HTML/PDF/DOCX to ODT)"

# Build function
build_image() {
    local dockerfile=$1
    local tag=$2
    local description=$3
    
    echo ""
    echo -e "${YELLOW}Building:${NC} $tag"
    echo -e "${BLUE}Description:${NC} $description"
    echo -e "${BLUE}Dockerfile:${NC} $dockerfile"
    echo ""
    
    if [ ! -f "$dockerfile" ]; then
        echo -e "${RED}✗${NC} Dockerfile not found: $dockerfile"
        return 1
    fi
    
    # Build image
    $RUNTIME build -f "$dockerfile" -t "$tag:latest" .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Successfully built: $tag:latest"
        
        # Show image size
        if [ "$RUNTIME" = "podman" ]; then
            SIZE=$($RUNTIME images --format "{{.Size}}" "$tag:latest")
        else
            SIZE=$($RUNTIME images --format "{{.Size}}" "$tag:latest")
        fi
        echo -e "${BLUE}Size:${NC} $SIZE"
        return 0
    else
        echo -e "${RED}✗${NC} Failed to build: $tag"
        return 1
    fi
}

# Build all images or specific one
if [ -n "$1" ]; then
    # Build specific image
    IMAGE_KEY="$1"
    if [ -z "${IMAGES[$IMAGE_KEY]}" ]; then
        echo -e "${RED}✗${NC} Unknown image: $IMAGE_KEY"
        echo ""
        echo "Available images:"
        for key in "${!IMAGES[@]}"; do
            IFS=':' read -r dockerfile tag description <<< "${IMAGES[$key]}"
            echo -e "  ${YELLOW}$key${NC} - $description"
        done
        exit 1
    fi
    
    IFS=':' read -r dockerfile tag description <<< "${IMAGES[$IMAGE_KEY]}"
    build_image "$dockerfile" "$tag" "$description"
else
    # Build all images
    echo "Building all images..."
    
    SUCCESS_COUNT=0
    FAIL_COUNT=0
    
    for key in "${!IMAGES[@]}"; do
        IFS=':' read -r dockerfile tag description <<< "${IMAGES[$key]}"
        if build_image "$dockerfile" "$tag" "$description"; then
            ((SUCCESS_COUNT++))
        else
            ((FAIL_COUNT++))
        fi
    done
    
    echo ""
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Build Summary${NC}"
    echo -e "${BLUE}================================${NC}"
    echo -e "${GREEN}Success:${NC} $SUCCESS_COUNT"
    if [ $FAIL_COUNT -gt 0 ]; then
        echo -e "${RED}Failed:${NC} $FAIL_COUNT"
    fi
fi

echo ""
echo -e "${GREEN}✓${NC} Done!"
echo ""
echo "Images built:"
echo "  • mcp-skills-docx             - DOCX only (defusedxml)"
echo "  • mcp-skills-pptx             - PPTX only (python-pptx, Pillow)"
echo "  • mcp-skills-xlsx             - XLSX only (openpyxl)"
echo "  • mcp-skills-xlsx-libreoffice - XLSX + formula recalc (openpyxl + LibreOffice)"
echo "  • mcp-skills-pdf              - PDF only (pypdf, pdf2image, OCR)"
echo "  • mcp-skills-office           - All Office formats (combined)"
echo "  • mcp-skills-document-parsing - Document parsing (pandoc, pdfplumber, lxml)"
echo ""
