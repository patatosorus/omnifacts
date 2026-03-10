#!/usr/bin/env bash
set -euo pipefail

# ─── Configuration ───────────────────────────────────────────────────────────

BASE_URL="${BASE_URL:-http://localhost:8080}"
API="$BASE_URL/api"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-password}"
DB_NAME="${DB_NAME:-postgres}"

PASS=0
FAIL=0
TOTAL=0

# ─── Helpers ─────────────────────────────────────────────────────────────────

red()   { printf "\033[31m%s\033[0m" "$*"; }
green() { printf "\033[32m%s\033[0m" "$*"; }
bold()  { printf "\033[1m%s\033[0m" "$*"; }
dim()   { printf "\033[2m%s\033[0m" "$*"; }

assert_status() {
    local label="$1" expected="$2" actual="$3"
    TOTAL=$((TOTAL + 1))
    if [ "$actual" -eq "$expected" ]; then
        PASS=$((PASS + 1))
        printf "  %-50s %s\n" "$label" "$(green "✓ $actual")"
    else
        FAIL=$((FAIL + 1))
        printf "  %-50s %s\n" "$label" "$(red "✗ $actual (attendu $expected)")"
    fi
}

assert_json() {
    local label="$1" jq_expr="$2" expected="$3" body="$4"
    TOTAL=$((TOTAL + 1))
    local actual
    actual=$(echo "$body" | jq -r "$jq_expr" 2>/dev/null || echo "PARSE_ERROR")
    if [ "$actual" = "$expected" ]; then
        PASS=$((PASS + 1))
        printf "  %-50s %s\n" "$label" "$(green "✓ $actual")"
    else
        FAIL=$((FAIL + 1))
        printf "  %-50s %s\n" "$label" "$(red "✗ $actual (attendu $expected)")"
    fi
}

api_get() {
    curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" "$API$1" 2>/dev/null || echo -e "\n000"
}

api_post_json() {
    curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$2" "$API$1" 2>/dev/null || echo -e "\n000"
}

api_put_json() {
    curl -s -w "\n%{http_code}" \
        -X PUT \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$2" "$API$1" 2>/dev/null || echo -e "\n000"
}

api_delete() {
    curl -s -w "\n%{http_code}" \
        -X DELETE \
        -H "Authorization: Bearer $TOKEN" \
        "$API$1" 2>/dev/null || echo -e "\n000"
}

api_upload() {
    curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer $TOKEN" \
        -F "file=@$2" \
        -F "name=$3" \
        -F "version=$4" \
        "$API$1" 2>/dev/null || echo -e "\n000"
}

split_response() {
    local raw="$1"
    BODY=$(echo "$raw" | head -n -1)
    HTTP_CODE=$(echo "$raw" | tail -n 1)
}

promote_admin() {
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -qc \
        "UPDATE users SET role = 'admin' WHERE username = '$1';" 2>/dev/null
}

# ─── Preflight ───────────────────────────────────────────────────────────────

for cmd in curl jq psql; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "$(red "Erreur:") $cmd est requis mais non trouvé"
        exit 1
    fi
done

echo ""
bold "══════════════════════════════════════════════════════════════"
echo ""
bold "  Omnifacts Smoke Test"
echo ""
bold "══════════════════════════════════════════════════════════════"
echo ""
dim "  Serveur: $BASE_URL"
echo ""

# ─── 1. Health ───────────────────────────────────────────────────────────────

bold "▸ Health"
echo ""
split_response "$(curl -s -w "\n%{http_code}" "$API/health" 2>/dev/null || echo -e "\n000")"
assert_status "GET /api/health" 200 "$HTTP_CODE"
assert_json   ".success = true" ".success" "true" "$BODY"
echo ""

# ─── 2. Auth ─────────────────────────────────────────────────────────────────

bold "▸ Authentification"
echo ""

TIMESTAMP=$(date +%s)
USERNAME="smoke-admin-$TIMESTAMP"
EMAIL="smoke-$TIMESTAMP@test.local"
PASSWORD="smoketest1234"

split_response "$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
    "$API/auth/register" 2>/dev/null || echo -e "\n000")"
assert_status "POST /api/auth/register" 201 "$HTTP_CODE"

split_response "$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
    "$API/auth/login" 2>/dev/null || echo -e "\n000")"
assert_status "POST /api/auth/login" 200 "$HTTP_CODE"
TOKEN=$(echo "$BODY" | jq -r '.data.token')
if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    assert_status "  token non vide" 0 0
else
    assert_status "  token non vide" 0 1
fi

promote_admin "$USERNAME"

split_response "$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
    "$API/auth/login" 2>/dev/null || echo -e "\n000")"
TOKEN=$(echo "$BODY" | jq -r '.data.token')

split_response "$(curl -s -w "\n%{http_code}" "$API/repos" 2>/dev/null || echo -e "\n000")"
assert_status "GET /api/repos sans auth → 401" 401 "$HTTP_CODE"

split_response "$(api_get "/repos")"
assert_status "GET /api/repos avec auth → 200" 200 "$HTTP_CODE"
echo ""

# ─── 3. Plugin types ────────────────────────────────────────────────────────

bold "▸ Plugins & types"
echo ""

split_response "$(api_get "/repos/types")"
assert_status "GET /api/repos/types" 200 "$HTTP_CODE"
assert_json   "  generic présent" '.data | map(select(. == "generic")) | length' "1" "$BODY"

TYPES_COUNT=$(echo "$BODY" | jq '.data | length')
printf "  %-50s %s\n" "  types chargés" "$(dim "$TYPES_COUNT")"
echo ""

# ─── 4. Repositories ────────────────────────────────────────────────────────

bold "▸ Dépôts"
echo ""

REPO_NAME="smoke-repo-$TIMESTAMP"

split_response "$(api_post_json "/repos/generic" "{\"name\":\"$REPO_NAME\",\"description\":\"Smoke test repo\"}")"
assert_status "POST /api/repos/generic (create)" 201 "$HTTP_CODE"

split_response "$(api_get "/repos/generic/$REPO_NAME")"
assert_status "GET /api/repos/generic/$REPO_NAME" 200 "$HTTP_CODE"
assert_json   "  nom correct" ".data.name" "$REPO_NAME" "$BODY"
assert_json   "  type = generic" ".data.artefact_type" "generic" "$BODY"

split_response "$(api_get "/repos")"
assert_status "GET /api/repos (list)" 200 "$HTTP_CODE"

split_response "$(api_get "/repos/docker/$REPO_NAME")"
assert_status "GET /api/repos/docker/$REPO_NAME → 404" 404 "$HTTP_CODE"

DOCKER_REPO="smoke-docker-$TIMESTAMP"
split_response "$(api_post_json "/repos/docker" "{\"name\":\"$DOCKER_REPO\"}")"
assert_status "POST /api/repos/docker (create)" 201 "$HTTP_CODE"
echo ""

# ─── 5. Artefact upload / download / delete ──────────────────────────────────

bold "▸ Artefacts (upload → list → download → delete)"
echo ""

TMPFILE=$(mktemp)
echo "Omnifacts smoke test payload — $(date -Iseconds)" > "$TMPFILE"
EXPECTED_CONTENT=$(cat "$TMPFILE")

split_response "$(api_upload "/repos/generic/$REPO_NAME/artefacts" "$TMPFILE" "smoke-artifact" "0.1.0")"
assert_status "POST upload artefact" 201 "$HTTP_CODE"
ARTEFACT_ID=$(echo "$BODY" | jq -r '.data.id')
printf "  %-50s %s\n" "  artefact id" "$(dim "$ARTEFACT_ID")"

split_response "$(api_get "/repos/generic/$REPO_NAME/artefacts")"
assert_status "GET list artefacts" 200 "$HTTP_CODE"
ARTEFACT_COUNT=$(echo "$BODY" | jq '.data | length')
assert_json   "  au moins 1 artefact" '.data | length > 0' "true" "$BODY"

DOWNLOADED=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$API/repos/generic/$REPO_NAME/artefacts/$ARTEFACT_ID/content" 2>/dev/null || echo "DOWNLOAD_FAILED")
TOTAL=$((TOTAL + 1))
if [ "$DOWNLOADED" = "$EXPECTED_CONTENT" ]; then
    PASS=$((PASS + 1))
    printf "  %-50s %s\n" "GET download contenu identique" "$(green "✓")"
else
    FAIL=$((FAIL + 1))
    printf "  %-50s %s\n" "GET download contenu identique" "$(red "✗ contenu différent")"
fi

split_response "$(api_delete "/repos/generic/$REPO_NAME/artefacts/$ARTEFACT_ID")"
assert_status "DELETE artefact" 200 "$HTTP_CODE"

split_response "$(api_get "/repos/generic/$REPO_NAME/artefacts")"
assert_json   "  artefact supprimé" '.data | length' "0" "$BODY"
echo ""

# ─── 6. Permissions ──────────────────────────────────────────────────────────

bold "▸ Permissions"
echo ""

USER2="smoke-user-$TIMESTAMP"
split_response "$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USER2\",\"email\":\"$USER2@test.local\",\"password\":\"$PASSWORD\"}" \
    "$API/auth/register" 2>/dev/null || echo -e "\n000")"
USER2_ID=$(echo "$BODY" | jq -r '.data.id')

split_response "$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USER2\",\"password\":\"$PASSWORD\"}" \
    "$API/auth/login" 2>/dev/null || echo -e "\n000")"
USER2_TOKEN=$(echo "$BODY" | jq -r '.data.token')

split_response "$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $USER2_TOKEN" \
    "$API/repos/generic/$REPO_NAME/artefacts" 2>/dev/null || echo -e "\n000")"
assert_status "GET artefacts sans permission → 403" 403 "$HTTP_CODE"

split_response "$(api_put_json "/repos/generic/$REPO_NAME/permissions" "{\"user_id\":\"$USER2_ID\",\"level\":\"read\"}")"
assert_status "PUT grant read" 200 "$HTTP_CODE"

split_response "$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $USER2_TOKEN" \
    "$API/repos/generic/$REPO_NAME/artefacts" 2>/dev/null || echo -e "\n000")"
assert_status "GET artefacts avec read → 200" 200 "$HTTP_CODE"
echo ""

# ─── 7. Cleanup ──────────────────────────────────────────────────────────────

bold "▸ Nettoyage"
echo ""

split_response "$(api_delete "/repos/generic/$REPO_NAME")"
assert_status "DELETE repo generic" 200 "$HTTP_CODE"

split_response "$(api_delete "/repos/docker/$DOCKER_REPO")"
assert_status "DELETE repo docker" 200 "$HTTP_CODE"

rm -f "$TMPFILE"
echo ""

# ─── Summary ─────────────────────────────────────────────────────────────────

bold "══════════════════════════════════════════════════════════════"
echo ""
if [ "$FAIL" -eq 0 ]; then
    printf "  %s  %s/%s tests passés\n" "$(green "PASS")" "$PASS" "$TOTAL"
else
    printf "  %s  %s/%s tests passés, %s échoués\n" "$(red "FAIL")" "$PASS" "$TOTAL" "$(red "$FAIL")"
fi
echo ""
bold "══════════════════════════════════════════════════════════════"
echo ""

exit "$FAIL"
