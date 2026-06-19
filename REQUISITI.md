# Stasera — Requisiti di progetto

> Documento di riferimento per Claude Code.
> Lingua del codice: **inglese**. Lingua UI app: **italiano**.
> Lingua commenti e commit: **italiano**.

---

## 1. Visione del progetto

**Stasera** è un'app mobile per single che lavorano a tempo pieno e hanno poco tempo
da dedicare alla cucina. Risolve un problema preciso: **eliminare la decisione quotidiana
su cosa mangiare la sera**.

Il loop settimanale è fisso e non ha varianti:

1. **Domenica (5 minuti)** — l'AI genera 5 cene per la settimana (lun–ven).
   L'utente può approvare o scambiare singoli giorni con un tap.
2. **Lista della spesa** — generata automaticamente dal piano pasti,
   organizzata per corsia del supermercato, spesa una volta sola.
3. **Ogni sera** — apri l'app, vedi la cena di stasera, cucini.
   Nessuna decisione, nessuna sorpresa.
4. **Modalità "troppo stanco"** — tasto di emergenza che propone
   3 opzioni da 10 minuti o meno usando solo prodotti staple (sempre in casa).

---

## 2. Stack tecnologico

### Backend — Go + Echo

| Componente | Libreria / Tool |
|---|---|
| Framework HTTP | `github.com/labstack/echo/v4` |
| JWT middleware | `github.com/labstack/echo-jwt/v4` |
| JWT signing | `github.com/golang-jwt/jwt/v5` |
| Database driver | `github.com/go-sql-driver/mysql` (via `database/sql` standard) |
| Password hashing | `golang.org/x/crypto/bcrypt` |
| Validazione input | `github.com/go-playground/validator/v10` |
| Env vars | `github.com/joho/godotenv` |
| UUID | `github.com/google/uuid` (generati in Go, non dal DB) |
| AI gateway | `github.com/sashabaranov/go-openai` (puntato su Vercel AI Gateway) |

**Nessun ORM.** Query SQL scritte a mano con `database/sql` + `go-sql-driver/mysql` direttamente.
Più verboso ma esplicito, performante e senza magia nascosta.

### Perché go-openai + Vercel AI Gateway

Il Vercel AI SDK è TypeScript-only — non esiste un SDK Go ufficiale Vercel.
Il Vercel AI Gateway espone però un endpoint **compatibile con le OpenAI API** a
`https://ai-gateway.vercel.sh/v1`. Si usa quindi `go-openai` (la libreria Go
più battle-tested per OpenAI, 10k+ star) con base URL custom che punta al gateway.

Vantaggi rispetto a chiamate HTTP native:
- SDK tipizzato con struct già pronte per request/response
- Gestione automatica di retry, timeout, stream
- Cambio modello in una riga (`anthropic/claude-sonnet-4-20250514` → `openai/gpt-4o`)
- Dashboard Vercel per monitoring, budget, usage per modello
- Nessuna API key Anthropic da gestire — solo la Vercel AI Gateway key

### Frontend — Flutter (mobile-first)

| Componente | Libreria |
|---|---|
| State management | `flutter_riverpod` + `riverpod_annotation` |
| Navigazione | `go_router` |
| HTTP client | `dio` |
| Storage sicuro (JWT) | `flutter_secure_storage` |
| Cache locale | `shared_preferences` (solo preferenze leggere) |
| Timer cucina | built-in `dart:async` |
| Animazione completamento | `lottie` |
| Notifiche locali | `flutter_local_notifications` |

**Target: solo mobile** (iOS + Android). Nessun supporto web richiesto.

### Infrastruttura locale

```
docker-compose:
  - mysql:8
  - (opzionale) phpMyAdmin / DBeaver per ispezione DB
```

---

## 3. Struttura progetto backend

```
stasera-api/
├── cmd/
│   └── main.go                  # Entry point: setup Echo, middleware, routes
├── internal/
│   ├── config/
│   │   └── config.go            # Lettura .env, struct Config
│   ├── db/
│   │   └── db.go                # sql.Open("mysql"), ping, close
│   ├── model/
│   │   ├── user.go
│   │   ├── meal_plan.go
│   │   ├── recipe.go
│   │   ├── shopping_list.go
│   │   └── staple.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── meal_plan_repo.go
│   │   ├── recipe_repo.go
│   │   ├── shopping_list_repo.go
│   │   └── staple_repo.go
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── meal_plan_handler.go
│   │   ├── recipe_handler.go
│   │   ├── shopping_list_handler.go
│   │   ├── staple_handler.go
│   │   └── ai_handler.go
│   ├── middleware/
│   │   └── auth.go              # Estrazione userID da JWT claims
│   ├── service/
│   │   ├── meal_plan_service.go
│   │   └── shopping_list_service.go
│   └── ai/
│       └── anthropic.go         # Client Anthropic API, prompt builder
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_recipes.sql
│   ├── 003_create_meal_plans.sql
│   ├── 004_create_shopping_lists.sql
│   └── 005_seed_staples.sql
├── .env
├── .env.example
├── go.mod
├── go.sum
└── docker-compose.yml
```

**Convenzione handlers:** ogni handler riceve `(c echo.Context)` e chiama il repository
o il service corrispondente. Nessuna logica di business negli handler.

---

## 4. Schema del database

Migrations SQL pure in `migrations/`. Eseguite in ordine all'avvio se non già presenti
(controllo manuale con tabella `schema_migrations`). Ogni file contiene una singola
istruzione `CREATE TABLE` (no split su `;`).

Le tabelle usano `CHAR(36)` come PK per gli UUID (generati in Go via `google/uuid`
e passati esplicitamente all'INSERT, non come DEFAULT del DB). I timestamp usano
`TIMESTAMP` (MySQL non ha `TIMESTAMPTZ`); il DSN deve includere `parseTime=true`
per mapparli su `time.Time`. Le colonne array di Postgres (`TEXT[]`) sono mappate
su `JSON` in MySQL per coerenza con `ingredients`/`steps`.

### users
```sql
CREATE TABLE users (
    id            CHAR(36) NOT NULL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name  VARCHAR(100) NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### recipes
```sql
-- Ricette generate dall'AI e cachate per utente.
-- Non ci sono ricette "globali": ogni utente ha le sue.
CREATE TABLE recipes (
    id              CHAR(36) NOT NULL PRIMARY KEY,
    user_id         CHAR(36) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    prep_minutes    SMALLINT NOT NULL,          -- max 30 per ricette normali
    servings        SMALLINT NOT NULL DEFAULT 1,
    ingredients     JSON NOT NULL,             -- [{ "name": "pasta", "qty": "80g" }]
    steps           JSON NOT NULL,             -- [{ "text": "...", "timer_seconds": 480 }]
    is_rescue       BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE = modalità troppo stanco
    times_cooked    SMALLINT NOT NULL DEFAULT 0,
    last_cooked_at  DATE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_recipes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### meal_plans
```sql
CREATE TABLE meal_plans (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    week_start  DATE NOT NULL,                  -- sempre lunedì
    status      VARCHAR(20) NOT NULL DEFAULT 'active', -- active | archived
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, week_start),
    CONSTRAINT fk_meal_plans_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### meal_plan_days
```sql
CREATE TABLE meal_plan_days (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    plan_id     CHAR(36) NOT NULL,
    day_of_week SMALLINT NOT NULL,              -- 1=lun, 2=mar ... 5=ven
    recipe_id   CHAR(36) NOT NULL,
    UNIQUE (plan_id, day_of_week),
    CONSTRAINT fk_meal_plan_days_plan FOREIGN KEY (plan_id) REFERENCES meal_plans(id) ON DELETE CASCADE,
    CONSTRAINT fk_meal_plan_days_recipe FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);
```

### shopping_lists
```sql
CREATE TABLE shopping_lists (
    id           CHAR(36) NOT NULL PRIMARY KEY,
    user_id      CHAR(36) NOT NULL,
    plan_id      CHAR(36),
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    CONSTRAINT fk_shopping_lists_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_shopping_lists_plan FOREIGN KEY (plan_id) REFERENCES meal_plans(id)
);
```

### shopping_items
```sql
CREATE TABLE shopping_items (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    list_id     CHAR(36) NOT NULL,
    name        VARCHAR(200) NOT NULL,
    quantity    VARCHAR(50) NOT NULL,           -- es. "80g", "1 scatoletta", "6 pz"
    aisle       VARCHAR(50) NOT NULL,           -- carne, dispensa, frigo, verdura, altro
    is_checked  BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order  SMALLINT NOT NULL DEFAULT 0,
    CONSTRAINT fk_shopping_items_list FOREIGN KEY (list_id) REFERENCES shopping_lists(id) ON DELETE CASCADE
);
```

### staples
```sql
-- Prodotti sempre in casa dell'utente (pasta, uova, olio, tonno...)
-- Usati dalla modalità "troppo stanco" per sapere cosa c'è sicuramente disponibile.
CREATE TABLE staples (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    name        VARCHAR(200) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (user_id, name),
    CONSTRAINT fk_staples_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### user_preferences
```sql
-- disliked_ingredients e preferred_cuisines sono colonne JSON (non TEXT[] come su Postgres)
-- perché MySQL non supporta array nativi. Si leggono/scrivono come []byte + json.Unmarshal.
CREATE TABLE user_preferences (
    user_id              CHAR(36) NOT NULL PRIMARY KEY,
    disliked_ingredients JSON NOT NULL,         -- es. ["fegato","trippa"]
    max_prep_minutes    SMALLINT NOT NULL DEFAULT 30,
    preferred_cuisines  JSON NOT NULL,          -- es. ["italiana","giapponese"]
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_preferences_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

## 5. Modelli Go

```go
// internal/model/user.go
type User struct {
    ID           uuid.UUID `db:"id"`
    Email        string    `db:"email"`
    PasswordHash string    `db:"password_hash"`
    DisplayName  string    `db:"display_name"`
    CreatedAt    time.Time `db:"created_at"`
}

// internal/model/recipe.go
type RecipeIngredient struct {
    Name string `json:"name"`
    Qty  string `json:"qty"`
}

type RecipeStep struct {
    Text         string `json:"text"`
    TimerSeconds int    `json:"timer_seconds,omitempty"` // 0 = nessun timer
}

type Recipe struct {
    ID           uuid.UUID          `db:"id"          json:"id"`
    UserID       uuid.UUID          `db:"user_id"     json:"user_id"`
    Name         string             `db:"name"        json:"name"`
    PrepMinutes  int                `db:"prep_minutes" json:"prep_minutes"`
    Servings     int                `db:"servings"    json:"servings"`
    Ingredients  []RecipeIngredient `db:"ingredients" json:"ingredients"` // MySQL JSON, letto via []byte + json.Unmarshal
    Steps        []RecipeStep       `db:"steps"       json:"steps"`
    IsRescue     bool               `db:"is_rescue"   json:"is_rescue"`
    TimesCooked  int                `db:"times_cooked" json:"times_cooked"`
    LastCookedAt *time.Time         `db:"last_cooked_at" json:"last_cooked_at,omitempty"`
    CreatedAt    time.Time          `db:"created_at"  json:"created_at"`
}

// internal/model/meal_plan.go
type MealPlan struct {
    ID        uuid.UUID    `db:"id"         json:"id"`
    UserID    uuid.UUID    `db:"user_id"    json:"user_id"`
    WeekStart time.Time    `db:"week_start" json:"week_start"`
    Status    string       `db:"status"     json:"status"`
    Days      []MealPlanDay `json:"days"`   // join, non colonna DB
    CreatedAt time.Time    `db:"created_at" json:"created_at"`
}

type MealPlanDay struct {
    ID        uuid.UUID `db:"id"          json:"id"`
    PlanID    uuid.UUID `db:"plan_id"     json:"plan_id"`
    DayOfWeek int       `db:"day_of_week" json:"day_of_week"` // 1–5
    RecipeID  uuid.UUID `db:"recipe_id"   json:"recipe_id"`
    Recipe    *Recipe   `json:"recipe,omitempty"`              // join opzionale
}

// internal/model/shopping_list.go
type ShoppingItem struct {
    ID        uuid.UUID `db:"id"         json:"id"`
    ListID    uuid.UUID `db:"list_id"    json:"list_id"`
    Name      string    `db:"name"       json:"name"`
    Quantity  string    `db:"quantity"   json:"quantity"`
    Aisle     string    `db:"aisle"      json:"aisle"`
    IsChecked bool      `db:"is_checked" json:"is_checked"`
    SortOrder int       `db:"sort_order" json:"sort_order"`
}

type ShoppingList struct {
    ID          uuid.UUID      `db:"id"           json:"id"`
    UserID      uuid.UUID      `db:"user_id"      json:"user_id"`
    PlanID      *uuid.UUID     `db:"plan_id"      json:"plan_id,omitempty"`
    Items       []ShoppingItem `json:"items"`
    CreatedAt   time.Time      `db:"created_at"   json:"created_at"`
    CompletedAt *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
}

// internal/model/staple.go
type Staple struct {
    ID       uuid.UUID `db:"id"        json:"id"`
    UserID   uuid.UUID `db:"user_id"   json:"user_id"`
    Name     string    `db:"name"      json:"name"`
    IsActive bool      `db:"is_active" json:"is_active"`
}
```

---

## 6. API Endpoints

Base URL: `/api/v1`
Tutti gli endpoint (tranne `/auth/*`) richiedono `Authorization: Bearer <token>`.

### Auth

```
POST   /api/v1/auth/register
       Body: { "email": string, "password": string, "display_name": string }
       → 201: { "user": UserDTO, "access_token": string, "refresh_token": string }

POST   /api/v1/auth/login
       Body: { "email": string, "password": string }
       → 200: { "user": UserDTO, "access_token": string, "refresh_token": string }

POST   /api/v1/auth/refresh
       Body: { "refresh_token": string }
       → 200: { "access_token": string, "refresh_token": string }

GET    /api/v1/auth/me
       → 200: UserDTO
```

JWT: access token expiry **15 minuti**, refresh token expiry **30 giorni**.
Entrambi firmati HS256 con `JWT_SECRET` da .env.

### Meal plan

```
GET    /api/v1/meal-plan/current
       # Piano della settimana corrente (se esiste)
       → 200: MealPlanDTO  |  404 se non esiste

POST   /api/v1/meal-plan/generate
       # Chiede all'AI di generare il piano per la prossima settimana lun–ven
       Body: {} (usa preferenze utente già salvate)
       → 201: MealPlanDTO

PATCH  /api/v1/meal-plan/:planId/days/:dayOfWeek
       # Scambia la ricetta di un giorno
       Body: { "recipe_id": uuid }  |  { "regenerate": true }
       → 200: MealPlanDayDTO con ricetta aggiornata

GET    /api/v1/meal-plan/today
       # Ricetta di stasera (basata sul giorno della settimana corrente)
       → 200: RecipeDTO  |  404 se oggi non è lun–ven o nessun piano attivo
```

### Recipes

```
GET    /api/v1/recipes
       # Tutte le ricette cachate dell'utente
       Query: ?is_rescue=true|false
       → 200: RecipeDTO[]

GET    /api/v1/recipes/:id
       → 200: RecipeDTO

POST   /api/v1/recipes/:id/cooked
       # Segna la ricetta come cucinata stasera (aggiorna times_cooked, last_cooked_at)
       → 200: RecipeDTO

DELETE /api/v1/recipes/:id
       # Rimuove una ricetta dalla cache (non usata in nessun piano attivo)
       → 204
```

### Shopping list

```
GET    /api/v1/shopping-list/current
       # Lista generata dall'ultimo piano settimanale
       → 200: ShoppingListDTO  |  404

POST   /api/v1/shopping-list/generate
       # Genera lista dalla meal plan corrente (sovrascrive la precedente se esiste)
       → 201: ShoppingListDTO

PATCH  /api/v1/shopping-list/items/:itemId
       Body: { "is_checked": boolean }
       → 200: ShoppingItemDTO

POST   /api/v1/shopping-list/current/complete
       # Marca la spesa come completata
       → 200: ShoppingListDTO
```

### Staples (prodotti sempre in casa)

```
GET    /api/v1/staples
       → 200: StapleDTO[]

POST   /api/v1/staples
       Body: { "name": string }
       → 201: StapleDTO

PATCH  /api/v1/staples/:id
       Body: { "is_active": boolean }
       → 200: StapleDTO

DELETE /api/v1/staples/:id
       → 204
```

### AI

```
POST   /api/v1/ai/rescue
       # Modalità "troppo stanco": genera 3 ricette da ≤10 min con gli staple attivi
       → 200: { "recipes": RecipeDTO[] }
       # Le ricette vengono salvate nel DB con is_rescue=true

GET    /api/v1/preferences
PATCH  /api/v1/preferences
       Body: { "disliked_ingredients": string[], "max_prep_minutes": int, "preferred_cuisines": string[] }
       → 200: PreferencesDTO
```

---

## 7. Integrazione AI — Vercel AI Gateway + go-openai

Il backend usa `github.com/sashabaranov/go-openai` con base URL puntata al
Vercel AI Gateway (`https://ai-gateway.vercel.sh/v1`).
La chiave API è in `.env` (`VERCEL_AI_GATEWAY_API_KEY`), mai nel client Flutter.
Modello default: `anthropic/claude-sonnet-4-20250514`.
Il modello è una stringa e può essere cambiato in un'unica variabile d'ambiente.

### Setup client Go

```go
// internal/ai/client.go
package ai

import (
    "github.com/sashabaranov/go-openai"
)

func NewGatewayClient(apiKey string) *openai.Client {
    cfg := openai.DefaultConfig(apiKey)
    cfg.BaseURL = "https://ai-gateway.vercel.sh/v1"
    return openai.NewClientWithConfig(cfg)
}
```

Utilizzo in `ai/anthropic.go` (rinominato `ai/gateway.go`):

```go
resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
    Model: cfg.AIModel,   // "anthropic/claude-sonnet-4-20250514"
    Messages: []openai.ChatCompletionMessage{
        {Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
        {Role: openai.ChatMessageRoleUser,   Content: userPrompt},
    },
    MaxTokens:   2000,
    Temperature: 0.7,
})
```

Max tokens: 2000. Timeout context: 30 secondi.

### 7.1 Generazione piano settimanale

**Endpoint chiamato:** `POST /api/v1/meal-plan/generate`

**Dati inviati al modello:**
- Preferenze utente (cibi non graditi, tempo massimo)
- Lista ricette già usate nelle ultime 2 settimane (per evitare ripetizioni)
- Giorno della settimana corrente

**System prompt:**
```
Sei un assistente culinario per una persona single che lavora a tempo pieno.
Devi suggerire 5 cene per la settimana (lunedì-venerdì).

Regole obbligatorie:
- Tempo di preparazione MASSIMO: {max_prep_minutes} minuti
- Ingredienti vietati: {disliked_ingredients}
- Non ripetere proteine consecutive (es. pollo lunedì e pollo martedì)
- Lunedì e martedì: ricette più semplici (pochi ingredienti, esecuzione meccanica)
- Venerdì: qualcosa di più soddisfacente, è fine settimana
- Porzioni per 1 persona
- Cucina preferita: {preferred_cuisines}
- NON suggerire ricette già usate di recente: {recent_recipes}

Rispondi SOLO con un array JSON valido, senza testo aggiuntivo, senza markdown.
Formato:
[
  {
    "day_of_week": 1,
    "name": "Nome ricetta",
    "prep_minutes": 20,
    "ingredients": [{"name": "pasta", "qty": "80g"}, ...],
    "steps": [{"text": "Descrizione passo", "timer_seconds": 480}, ...]
  },
  ...
]
```

### 7.2 Modalità "troppo stanco"

**Endpoint chiamato:** `POST /api/v1/ai/rescue`

**System prompt:**
```
L'utente è esausto dopo una lunga giornata lavorativa.
Ha in casa solo questi prodotti: {staples_list}

Suggerisci esattamente 3 cene di emergenza con queste regole:
- Tempo MASSIMO 10 minuti
- SOLO ingredienti dalla lista fornita, nessun altro
- Ordinate dal più semplice al meno semplice
- La terza può essere anche fredda/senza cottura

Rispondi SOLO con un array JSON valido, senza testo aggiuntivo, senza markdown.
Stesso formato del piano settimanale, senza il campo day_of_week.
```

### 7.3 Generazione lista della spesa da piano

Questa funzione non usa l'AI: è logica pura Go.
Il service `shopping_list_service.go` aggrega gli ingredienti di tutte e 5 le ricette
del piano, raggruppa per corsia (`aisle`), deduplica e somma le quantità dello stesso
ingrediente. Le corsie sono:
- `carne` — carne, pesce, salumi
- `frigo` — latticini, uova, prodotti freschi
- `verdura` — frutta e verdura
- `dispensa` — pasta, riso, conserve, olio, scatolette
- `altro` — tutto il resto

---

## 8. Struttura Flutter

```
lib/
├── main.dart
├── app.dart                      # MaterialApp con go_router
├── core/
│   ├── api/
│   │   ├── api_client.dart       # Dio instance, base URL, interceptors JWT
│   │   └── api_exception.dart
│   ├── auth/
│   │   └── auth_storage.dart     # flutter_secure_storage wrapper
│   └── theme/
│       └── app_theme.dart
├── features/
│   ├── auth/
│   │   ├── data/auth_repository.dart
│   │   ├── providers/auth_provider.dart
│   │   └── presentation/
│   │       ├── login_screen.dart
│   │       └── register_screen.dart
│   ├── home/
│   │   ├── data/meal_plan_repository.dart
│   │   ├── providers/tonight_provider.dart
│   │   └── presentation/
│   │       └── home_screen.dart       # Schermata principale: cena di stasera
│   ├── cooking/
│   │   ├── providers/cooking_provider.dart
│   │   └── presentation/
│   │       └── cooking_screen.dart    # Step-by-step con timer
│   ├── week_plan/
│   │   ├── data/meal_plan_repository.dart
│   │   ├── providers/week_plan_provider.dart
│   │   └── presentation/
│   │       └── week_plan_screen.dart  # Pianificatore domenicale
│   ├── shopping/
│   │   ├── data/shopping_repository.dart
│   │   ├── providers/shopping_provider.dart
│   │   └── presentation/
│   │       └── shopping_screen.dart   # Lista della spesa con checkboxes
│   ├── rescue/
│   │   ├── providers/rescue_provider.dart
│   │   └── presentation/
│   │       └── rescue_screen.dart     # Modalità troppo stanco
│   └── settings/
│       └── presentation/
│           └── settings_screen.dart   # Preferenze, staple, account
└── shared/
    └── widgets/
        ├── recipe_card.dart
        ├── step_card.dart
        └── loading_overlay.dart
```

### Navigazione (go_router)

```
/                     → HomeScreen (cena di stasera)
/cooking/:recipeId    → CookingScreen
/week                 → WeekPlanScreen
/shopping             → ShoppingScreen
/rescue               → RescueScreen
/settings             → SettingsScreen
/login                → LoginScreen (redirect se non auth)
/register             → RegisterScreen
```

`BottomNavigationBar` con 4 tab: Home · Settimana · Spesa · Impostazioni.

---

## 9. Schermate Flutter — dettaglio

### HomeScreen (schermata principale)

Aperta ogni sera dopo il lavoro. Deve essere **istantanea e chiarissima**.

- In alto: giorno della settimana + data
- Card grande centrale: nome ricetta, tempo preparazione, numero ingredienti
- Pulsante primario grande: "Inizia a cucinare" → naviga a `/cooking/:recipeId`
- Pulsante secondario arancione: "Troppo stanco — cambia" → naviga a `/rescue`
- Se oggi non è lun–ven o non c'è un piano attivo: messaggio "Nessuna cena pianificata
  per oggi. Vai su Settimana per generare il piano domenica."

**Stato empty state domenica:** messaggio "È domenica! Pianifica le cene della settimana."
con pulsante che apre `WeekPlanScreen`.

### CookingScreen

- Header: nome ricetta + timer totale (countdown dall'apertura)
- Sezione ingredienti collassabile (già li hai, non servono durante la cottura)
- Lista step con numero, testo, timer individuale per step (se `timer_seconds > 0`)
- Tap su uno step → mostra timer countdown in overlay, suona quando finisce
- Step completato: tap → si barra, passa automaticamente al successivo
- Al completamento di tutti gli step: animazione Lottie confetti + chiamata
  `POST /recipes/:id/cooked`

### WeekPlanScreen

- Lista lun–ven, ognuno con nome ricetta e tempo preparazione
- Tap su un giorno: bottom sheet con opzioni "Rigenera questo giorno" / "Scegli ricetta"
- Pulsante "Genera piano con AI" (solo se non esiste piano per la settimana corrente)
  → mostra loading con messaggio "L'AI sta pensando al tuo menù..."
  → mostra il piano generato, l'utente approva o modifica
- Pulsante "Genera lista della spesa" (dopo aver approvato il piano)

### ShoppingScreen

- Header: nome lista + stato (attiva / completata)
- Items raggruppati per corsia con header di sezione
- Checkbox per ogni item, swipe per spuntare
- Barra progresso in cima (N/TOT items)
- Pulsante "Spesa completata" quando tutti gli items sono spuntati

### RescueScreen

- Titolo: "Stasera cucino poco"
- Loading: "Sto cercando cosa hai in casa..." (2–3 secondi, chiamata AI)
- 3 card con ricette di emergenza: nome, tempo, ingredienti richiesti
- Tap su una card → va a CookingScreen

### SettingsScreen

- Sezione "Le mie preferenze": cibi da evitare (chip aggiungi/rimuovi), tempo max
- Sezione "Prodotti sempre in casa" (staple): lista con toggle on/off, aggiungi manuale
- Sezione account: display name, email, cambio password, logout

---

## 10. Gestione JWT in Flutter

`AuthInterceptor` aggiunto al client Dio:

```dart
// Aggiunge Bearer token a ogni richiesta
// Se riceve 401, prova refresh automatico
// Se il refresh fallisce, naviga a /login
```

Token salvati in `flutter_secure_storage`:
- `access_token` — JWT 15 minuti
- `refresh_token` — JWT 30 giorni

---

## 11. File .env backend

```env
# Server
PORT=8080

# Database
DATABASE_URL=stasera:stasera@tcp(localhost:3306)/stasera?parseTime=true&allowNativePasswords=true

# JWT
JWT_SECRET=<stringa-random-256-bit>
JWT_ACCESS_EXPIRY_MINUTES=15
JWT_REFRESH_EXPIRY_DAYS=30

# Vercel AI Gateway
# Crea la key su: https://vercel.com/dashboard → AI Gateway → API Keys
VERCEL_AI_GATEWAY_API_KEY=...
AI_MODEL=anthropic/claude-sonnet-4-20250514
# Cambia modello senza toccare il codice:
# AI_MODEL=openai/gpt-4o
# AI_MODEL=google/gemini-2.5-flash
```

---

## 12. docker-compose.yml

```yaml
services:
  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: stasera
      MYSQL_USER: stasera
      MYSQL_PASSWORD: stasera
    ports:
      - "3306:3306"
    volumes:
      - mysqldata:/var/lib/mysql

volumes:
  mysqldata:
```

---

## 13. Staple predefiniti (seed)

Al momento della registrazione, ogni utente riceve automaticamente questi staple:

```
pasta, riso, olio extravergine d'oliva, sale, pepe, aglio,
tonno in scatoletta, pomodori pelati in scatola, uova,
pane in cassetta, aceto, dado da brodo
```

L'utente può disattivarli o aggiungerne altri dal SettingsScreen.

---

## 14. Fasi di sviluppo

### Fase 1 — Backend base (giorni 1-3)
- [ ] Setup progetto Go: `go mod init`, struttura cartelle
- [ ] `docker-compose up` con PostgreSQL
- [ ] Migrations SQL: tutte e 5 le tabelle
- [ ] Config da .env con `godotenv`
- [ ] `db.go`: connessione `sql.Open("mysql", ...)`, ping all'avvio
- [ ] Modelli Go con tag `db` e `json`
- [ ] Repository: `user_repo.go` (Create, FindByEmail, FindByID)
- [ ] Handler: `auth_handler.go` (register, login, refresh, me)
- [ ] Middleware JWT con `echo-jwt/v4`
- [ ] Seed staple predefiniti alla registrazione
- [ ] Test manuale con curl / Postman

### Fase 2 — Meal plan + AI (giorni 4-6)
- [ ] Repository: `recipe_repo.go`, `meal_plan_repo.go`
- [ ] `ai/anthropic.go`: client HTTP, `GenerateMealPlan()`, `GenerateRescueMeals()`
- [ ] Service: `meal_plan_service.go` (orchestrazione AI → salvataggio ricette → piano)
- [ ] Handler: `meal_plan_handler.go`, `ai_handler.go`
- [ ] Test generazione piano con curl

### Fase 3 — Shopping list + Preferences (giorno 7)
- [ ] Repository: `shopping_list_repo.go`, `staple_repo.go`
- [ ] Service: `shopping_list_service.go` (aggregazione ingredienti, categorizzazione corsia)
- [ ] Handler: `shopping_list_handler.go`, `staple_handler.go`, preferences
- [ ] Handler: `recipe_handler.go` (GET list, GET detail, POST cooked)

### Fase 4 — Flutter base (giorni 8-10)
- [ ] Setup Flutter: Riverpod, go_router, Dio
- [ ] `AuthInterceptor` con refresh automatico
- [ ] Schermate auth: LoginScreen, RegisterScreen
- [ ] HomeScreen con chiamata `GET /meal-plan/today`
- [ ] WeekPlanScreen con lista giorni + pulsante genera

### Fase 5 — Flutter completo (giorni 11-14)
- [ ] CookingScreen con step-by-step e timer per step
- [ ] ShoppingScreen con checkbox e raggruppamento per corsia
- [ ] RescueScreen con chiamata AI e 3 ricette emergenza
- [ ] SettingsScreen: preferenze + gestione staple
- [ ] Animazione Lottie al completamento cottura
- [ ] Notifica locale: "Cosa mangi stasera?" ogni giorno feriale alle 18:30
- [ ] Test su dispositivo fisico Android e iOS

---

## 15. Note per Claude Code

- DB: MySQL 8 via `database/sql` + `github.com/go-sql-driver/mysql` (NO pgx, NO GORM).
- Il DSN in `.env` deve includere `parseTime=true` per mappare TIMESTAMP/DATETIME
  su `time.Time`. Placeholder query: `?` posizionali (non `$1`).
- Gli UUID sono `CHAR(36)` e vengono generati in Go con `uuid.NewString()` PRIMA
  dell'INSERT, poi passati come parametro. Non usare `RETURNING` (non supportato da
  MySQL): dopo l'INSERT fare una SELECT by id, oppure leggere `LastInsertId()`
  (inutilizzabile con UUID generato lato app).
- I campi JSON (`ingredients`, `steps`, `disliked_ingredients`, `preferred_cuisines`)
  si scrivono come stringa JSON (`json.Marshal`) e si leggono come `[]byte` con
  `json.Unmarshal`. Includere helper `scanRecipe()` / `scanUser()` nei repository.
- UNIQUE violation: l'errore MySQL è `*mysql.MySQLError` con `Number == 1062`
  (non il codice Postgres 23505). Helper `isDuplicateEntry(err)` in `repository/errors.go`.
- `ON CONFLICT ... DO NOTHING/UPDATE` di Postgres → `INSERT IGNORE` /
  `ON DUPLICATE KEY UPDATE col = VALUES(col)` in MySQL.
- Le colonne `TEXT[]` di Postgres non esistono in MySQL: mappate su `JSON` con
  marshalling esplicito nel repository.
- Echo: registrare le route protette in un gruppo `api := e.Group("/api/v1", jwtMiddleware)`.
- Il middleware `auth.go` deve estrarre `userID` dal JWT claims e metterlo nel
  `echo.Context` con `c.Set("userID", uid)` per averlo disponibile negli handler.
- Per le chiamate AI usare `github.com/sashabaranov/go-openai` con `cfg.BaseURL = "https://ai-gateway.vercel.sh/v1"` e `cfg.APIKey = VERCEL_AI_GATEWAY_API_KEY`. Il modello si specifica come stringa nel formato `creator/model-name` (es. `"anthropic/claude-sonnet-4-20250514"`). Creare un `AIGatewayClient` singleton inizializzato in `main.go` e iniettato negli handler che ne hanno bisogno.
- In Flutter: usare `ref.invalidate(tonightProvider)` dopo `POST /recipes/:id/cooked`
  per forzare il refresh della home.
- Non implementare WebSocket: l'app è mono-utente, nessuna sincronizzazione real-time.
- Notifiche locali Flutter: schedulare con `flutter_local_notifications` al login,
  riprogrammarle ogni volta che l'utente apre l'app.
- I timer nel CookingScreen usano `dart:async` `Timer.periodic` con stato locale
  in un `StateNotifier` — nessuna chiamata API durante la cottura.
