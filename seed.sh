#!/bin/bash
# Seed script for POS API - inserts dummy data
# Usage: bash seed.sh [API_URL]
# Default API_URL: http://localhost:3000/api/v1

API_URL="${1:-http://localhost:3000/api/v1}"

echo "================================================"
echo "  POS API Seed Script"
echo "  Target: $API_URL"
echo "================================================"
echo ""

# ========================================
# 1. Register Admin User
# ========================================
echo ">>> Registering admin user..."
curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "full_name": "Administrator"
  }' | python3 -m json.tool 2>/dev/null || echo "(may already exist)"

echo ""

# ========================================
# 2. Register Manager User
# ========================================
echo ">>> Registering manager user..."
curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "manager1",
    "password": "manager123",
    "full_name": "Budi Santoso"
  }' | python3 -m json.tool 2>/dev/null || echo "(may already exist)"

echo ""

# ========================================
# 3. Register Cashier User
# ========================================
echo ">>> Registering cashier user..."
curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "cashier1",
    "password": "cashier123",
    "full_name": "Siti Aminah"
  }' | python3 -m json.tool 2>/dev/null || echo "(may already exist)"

echo ""

# ========================================
# 4. Login as admin to get token
# ========================================
echo ">>> Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "ERROR: Could not get auth token. Is the API running?"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "Token obtained: ${TOKEN:0:20}..."
echo ""

AUTH="Authorization: Bearer $TOKEN"

# ========================================
# 5. Create Categories
# ========================================
echo ">>> Creating categories..."

CATEGORIES=("Makanan" "Minuman" "Snack" "Rokok" "Alat Tulis" "Peralatan Mandi" "Bumbu Masak")

declare -A CATEGORY_IDS

for CAT_NAME in "${CATEGORIES[@]}"; do
  RESPONSE=$(curl -s -X POST "$API_URL/categories" \
    -H "Content-Type: application/json" \
    -H "$AUTH" \
    -d "{\"name\": \"$CAT_NAME\"}")
  
  CAT_ID=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
  CATEGORY_IDS["$CAT_NAME"]=$CAT_ID
  echo "  Created: $CAT_NAME (ID: $CAT_ID)"
done

echo ""

# ========================================
# 6. Create Products
# ========================================
echo ">>> Creating products..."

# Makanan
create_product() {
  local name="$1" sku="$2" price="$3" cost="$4" stock="$5" cat_id="$6" desc="$7"
  RESPONSE=$(curl -s -X POST "$API_URL/products" \
    -H "Content-Type: application/json" \
    -H "$AUTH" \
    -d "{
      \"name\": \"$name\",
      \"sku\": \"$sku\",
      \"price\": $price,
      \"cost\": $cost,
      \"stock\": $stock,
      \"category_id\": $cat_id,
      \"description\": \"$desc\"
    }")
  local pid=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
  echo "  Created: $name (ID: $pid)"
}

# Get category IDs
CAT_MAKANAN=${CATEGORY_IDS["Makanan"]}
CAT_MINUMAN=${CATEGORY_IDS["Minuman"]}
CAT_SNACK=${CATEGORY_IDS["Snack"]}
CAT_ROKOK=${CATEGORY_IDS["Rokok"]}
CAT_ATK=${CATEGORY_IDS["Alat Tulis"]}
CAT_MANDI=${CATEGORY_IDS["Peralatan Mandi"]}
CAT_BUMBU=${CATEGORY_IDS["Bumbu Masak"]}

# Makanan products
create_product "Nasi Goreng Spesial"  "MKN-001" 25000 15000 100 "$CAT_MAKANAN" "Nasi goreng dengan telur dan ayam"
create_product "Mie Goreng Instan"    "MKN-002" 5000  3500  200 "$CAT_MAKANAN" "Mie goreng instan rasa ayam bawang"
create_product "Roti Tawar"           "MKN-003" 15000 10000 50  "$CAT_MAKANAN" "Roti tawar putih 400g"
create_product "Telur Ayam (6 pcs)"   "MKN-004" 12000 9000  80  "$CAT_MAKANAN" "Telur ayam ras 6 butir"
create_product "Sarden Kaleng"        "MKN-005" 18000 13000 40  "$CAT_MAKANAN" "Sarden kaleng saus tomat"

# Minuman products
create_product "Teh Botol Sosro"      "MNM-001" 5000  3000  150 "$CAT_MINUMAN" "Teh botol original 450ml"
create_product "Aqua 600ml"           "MNM-002" 4000  2500  300 "$CAT_MINUMAN" "Air mineral 600ml"
create_product "Kopi Good Day"        "MNM-003" 3000  2000  200 "$CAT_MINUMAN" "Kopi sachet instant"
create_product "Susu Ultra 250ml"     "MNM-004" 6000  4000  100 "$CAT_MINUMAN" "Susu UHT full cream"
create_product "Coca Cola 390ml"      "MNM-005" 7000  5000  80  "$CAT_MINUMAN" "Coca Cola original 390ml"
create_product "Sprite 390ml"         "MNM-006" 7000  5000  70  "$CAT_MINUMAN" "Sprite lemon lime 390ml"

# Snack products
create_product "Chitato 68g"          "SNK-001" 10000 7000  100 "$CAT_SNACK" "Chitato rasa sapi panggang"
create_product "Oreo Original"        "SNK-002" 8000  5500  90  "$CAT_SNACK" "Biskuit Oreo original 137g"
create_product "Coklat Silverqueen"   "SNK-003" 12000 8000  60  "$CAT_SNACK" "Silverqueen cashew 25g"
create_product "Beng-Beng"            "SNK-004" 3000  2000  150 "$CAT_SNACK" "Wafer coklat caramel"
create_product "Tango Wafer"          "SNK-005" 6000  4000  80  "$CAT_SNACK" "Tango wafer rasa keju 76g"

# Rokok products
create_product "Gudang Garam Surya"   "RKK-001" 28000 24000 50  "$CAT_ROKOK" "Gudang Garam Surya 12 batang"
create_product "Sampoerna Mild"       "RKK-002" 26000 22000 40  "$CAT_ROKOK" "Sampoerna A Mild 16 batang"
create_product "Dji Sam Soe"          "RKK-003" 30000 26000 30  "$CAT_ROKOK" "Dji Sam Soe 12 batang"

# Alat Tulis products
create_product "Pulpen Pilot"         "ATK-001" 5000  3000  100 "$CAT_ATK" "Pulpen Pilot BP-S 0.7mm"
create_product "Buku Tulis Sidu A5"   "ATK-002" 4000  2500  120 "$CAT_ATK" "Buku tulis Sinar Dunia 38 lembar"
create_product "Penghapus Joyko"      "ATK-003" 2000  1000  80  "$CAT_ATK" "Penghapus karet Joyko"

# Peralatan Mandi
create_product "Sabun Lifebuoy"       "MND-001" 3500  2500  100 "$CAT_MANDI" "Sabun batang Lifebuoy antibakteri"
create_product "Shampoo Sunsilk 70ml" "MND-002" 8000  5500  60  "$CAT_MANDI" "Shampoo Sunsilk hitam 70ml"
create_product "Pasta Gigi Pepsodent" "MND-003" 7000  5000  90  "$CAT_MANDI" "Pepsodent pencegah gigi berlubang 75g"

# Bumbu Masak
create_product "Kecap Bango 135ml"    "BMB-001" 8000  6000  70  "$CAT_BUMBU" "Kecap manis Bango 135ml"
create_product "Minyak Goreng 1L"     "BMB-002" 18000 15000 50  "$CAT_BUMBU" "Minyak goreng sawit 1 liter"
create_product "Gula Pasir 1kg"       "BMB-003" 16000 13000 40  "$CAT_BUMBU" "Gula pasir putih 1 kg"

echo ""
echo "================================================"
echo "  Seed Complete!"
echo "================================================"
echo ""
echo "Users created:"
echo "  admin    / admin123    (role: admin)"
echo "  manager1 / manager123  (role: admin*)"
echo "  cashier1 / cashier123  (role: admin*)"
echo ""
echo "* Note: All users register with 'admin' role by default."
echo "  To set correct roles, run the SQL commands below:"
echo ""
echo "  UPDATE users SET role='manager' WHERE username='manager1';"
echo "  UPDATE users SET role='cashier' WHERE username='cashier1';"
echo ""
echo "Run with: psql -U postgres -d pos -c \"UPDATE users SET role='manager' WHERE username='manager1'; UPDATE users SET role='cashier' WHERE username='cashier1';\""
echo ""
