#!/bin/bash

echo "🔗 URL Shortener Demo"
echo "===================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${BLUE}1. Testing health endpoint...${NC}"
curl -s "$BASE_URL/health" | jq .
echo ""

echo -e "${BLUE}2. Shortening a URL...${NC}"
echo "Creating short URL for: https://github.com/alexnthnz/url-shortener"

RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/alexnthnz/url-shortener"}')

echo "$RESPONSE" | jq .
SHORT_CODE=$(echo "$RESPONSE" | jq -r '.short_code')
echo ""

echo -e "${BLUE}3. Creating custom alias...${NC}"
echo "Creating custom alias 'my-demo-link' for: https://example.com"

CUSTOM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "custom_alias": "my-demo-link"}')

echo "$CUSTOM_RESPONSE" | jq .
echo ""

echo -e "${BLUE}4. Testing redirect (will show headers only)...${NC}"
echo "Testing redirect for short code: $SHORT_CODE"
curl -I -s "$BASE_URL/$SHORT_CODE"
echo ""

echo -e "${BLUE}5. Testing custom alias redirect...${NC}"
echo "Testing redirect for custom alias: my-demo-link"
curl -I -s "$BASE_URL/my-demo-link"
echo ""

echo -e "${BLUE}6. Getting URL statistics...${NC}"
echo "Statistics for short code: $SHORT_CODE"
curl -s "$BASE_URL/api/v1/urls/$SHORT_CODE/stats" | jq .
echo ""

echo "Statistics for custom alias: my-demo-link"
curl -s "$BASE_URL/api/v1/urls/my-demo-link/stats" | jq .
echo ""

echo -e "${GREEN}✅ Demo completed!${NC}"
echo ""
echo -e "${YELLOW}Note: Make sure the URL shortener server is running on $BASE_URL${NC}"
echo -e "${YELLOW}Start it with: make run${NC}" 