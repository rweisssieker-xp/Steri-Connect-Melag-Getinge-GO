# PRODUCT REQUIREMENTS DOCUMENT (PRD)

Software-Integration für Melag Cliniclave 45 & Getinge Aquadis 56 über Steri-Suite + GO-Schnittstelle

**Version:** 1.2  
**Status:** Validated - Complete with numbered FRs, API schemas, authentication  
**Date:** 2025-11-21  
**Zielgruppe:** CIO, IT-Leitung, Softwarearchitektur, Entwickler, Qualitätssicherung, Medizintechnik-Verantwortliche

---

## 1. Zielsetzung der Software

Die Software soll:

- Sterilisations- und Reinigungsprozesse von Medizintechnikgeräten in Praxis/Klinik vollständig digital unterstützen.
- Eine einheitliche Bedienoberfläche (Steri-Suite) für unterschiedliche Gerätetypen bereitstellen.
- Über eine lokale Middleware (GO-Schnittstelle) die Kommunikation zwischen Suite und Geräten übernehmen.
- Eine lückenlose Dokumentation, Traceability und Auditfähigkeit erfüllen.
- Melag-Geräte vollständig integrieren (Start → Status → Ergebnisprotokoll).
- Getinge-Geräte aufgrund der herstellerseitigen Einschränkungen zunächst nur als Online/Offline überwachen, mit Perspektive auf spätere Vollintegration.
- Eine lokale, portable Lösung ohne komplexe Installation sicherstellen (SQLite, portable App).
- **NEU:** Eine einfache Testoberfläche für Entwicklung, Debugging und Systemdiagnose bereitstellen.

---

## 2. Systemübersicht (Architektur)

Das System besteht aus drei Ebenen:

### 2.1 Steri-Suite (Frontend / Betriebssystemunabhängiges UI)

- Läuft im Browser (Electron/Chromium/Web-App).
- Zeigt Geräte, Status, Beladungsinformationen und Protokolle.
- Bietet Bedienung zur Prozessinitialisierung.
- Kommuniziert ausschließlich mit der GO-Schnittstelle per lokaler HTTP-API/WebSocket.

### 2.2 GO-Schnittstelle (Middleware / Backend)

- Lokaler Backend-Dienst als portable executable (Windows/Linux).
- Verarbeitet alle Befehle der Steri-Suite.
- Kommuniziert mit Melag und perspektivisch mit Getinge.
- Speichert Daten vollständig lokal in SQLite.
- Enthält Gerätetreiber/Adapter-Module.
- **NEU:** Bietet eine einfache Web-basierte Testoberfläche für Entwicklung und Debugging.

### 2.3 Geräte

**Melag Cliniclave 45**
- Sterilisator (Dampfautoklav)
- Liefert Start/Status/Ergebnisdaten
- Integration über dokumentierte oder freigegebene Melag-Schnittstelle (Format abhängig von Melag-Dokumentation)

**Getinge Aquadis 56**
- RDG (Reinigung + Desinfektion)
- Aktuell: nur ICMP erreichbar → kein API-Zugriff
- Geplante zukünftige Integration nur mit Herstellerfreigabe

---

## 3. Detaillierter Gesamtprozessablauf

### 3.1 Vorbereitung

Administrator öffnet die Steri-Suite.

Devices werden automatisch erkannt (falls möglich) oder manuell angelegt:
- Name
- Typ (Melag/ RDG)
- IP-Adresse
- Seriennummer
- Standort (optional)

GO-Schnittstelle speichert Geräteinformationen in SQLite.

### 3.2 Beladungsprozess (optional erweiterbar)

Anwender identifiziert Trays/Instrumente (manuell oder Barcode).

Suite speichert Beladung in SQLite → pending_load.

### 3.3 Start eines Melag-Zyklus (Sollprozess)

Nutzer wählt Gerät Melag Cliniclave 45 aus.

UI zeigt:
- aktueller Zustand (Bereit/Läuft/Fertig)
- letzte Zyklen

Nutzer klickt „Prozess Start".

Steri-Suite sendet API-Call an GO-Schnittstelle:
- POST /melag/{deviceId}/start

GO-Schnittstelle:
- Prüft Erreichbarkeit (Ping)
- Öffnet Melag-Verbindung (Typ abhängig von Melag-Dokumentation)
- Sendet Startsignal

Melag-Gerät startet Zyklus.

GO-Schnittstelle erhält Statusmeldungen:
- Phase (Aufheizen, Sterilisation, Trocknung)
- Parameter (Temperatur, Druck – falls verfügbar)
- Restzeit / Fortschritt

Steri-Suite zeigt laufenden Prozess an (WebSocket-Live-Stream).

Zyklusende:
- Gerät liefert Ergebnisdaten (OK/NOK, Programmart, Timestamp).
- GO-Schnittstelle schreibt alles in SQLite.
- UI zeigt „Zyklus abgeschlossen".

### 3.4 Getinge Aquadis 56 – aktueller Minimalprozess

Da keine offizielle API verfügbar:

GO-Schnittstelle führt zyklisch (alle 10–30 Sekunden) ICMP-Ping durch.

In SQLite wird gespeichert:
- Timestamp
- Reachable (0/1)

Steri-Suite zeigt:
- Gerät erreichbar (grün)
- Gerät nicht erreichbar (rot)

Keine weiteren Funktionen möglich, bis der Hersteller Zugang freigibt.

### 3.5 Protokollerstellung

Nach Prozessende erstellt die GO-Schnittstelle ein vollständiges Protokoll:
- Zyklus-ID
- Gerätedaten
- Startzeit / Endzeit
- Programmart
- Ergebnis
- Fehlercodes
- Verknüpfte Beladung (falls aktiviert)

Exportformate:
- PDF
- CSV
- JSON
- Auditreport (komplett)

### 3.6 Testoberfläche (NEU)

Die GO-Schnittstelle bietet eine einfache Web-basierte Testoberfläche (localhost) für:
- Entwicklung und Debugging
- Systemdiagnose
- API-Testing ohne Steri-Suite
- Datenbankinspektion
- Log-Viewing

---

## 4. Funktionale Anforderungen

**Scope Markierung:**
- **[MVP]** = Phase 1 (Must Have)
- **[Growth]** = Phase 2 (Should Have)
- **[Vision]** = Phase 3 (Future)

### 4.1 Anforderungen an Steri-Suite

#### Geräteverwaltung

**FR-001: [MVP] Geräte hinzufügen**
- Die Steri-Suite soll Geräte zur GO-Schnittstelle hinzufügen können.
- Erforderliche Informationen: Name, Typ (Melag/Getinge), IP-Adresse, Seriennummer, Standort (optional).

**FR-002: [MVP] Geräte bearbeiten**
- Die Steri-Suite soll bestehende Gerätekonfigurationen bearbeiten können.

**FR-003: [MVP] Geräte löschen**
- Die Steri-Suite soll Geräte aus der Konfiguration entfernen können.

**FR-004: [MVP] Gesundheitsmonitoring**
- Die Steri-Suite soll den Gesundheitsstatus (Ping, Verbindungsstatus) aller Geräte anzeigen.
- Status-Updates sollen in Echtzeit verfügbar sein.

#### Zyklussteuerung (nur Melag)

**FR-005: [MVP] Zyklus starten**
- Die Steri-Suite soll Sterilisationszyklen für Melag-Geräte starten können.
- Start-Befehl wird an GO-Schnittstelle gesendet.

**FR-006: [MVP] Laufende Zyklen anzeigen**
- Die Steri-Suite soll alle aktuell laufenden Zyklen anzeigen.

**FR-007: [MVP] Echtzeitstatus**
- Die Steri-Suite soll Echtzeitstatus-Updates für laufende Zyklen empfangen (WebSocket).
- Status soll Phase, Parameter (Temperatur, Druck), Restzeit enthalten.

#### Protokollverwaltung

**FR-008: [MVP] Zyklus-Listenansicht**
- Die Steri-Suite soll eine Listenansicht aller abgeschlossenen Zyklen bereitstellen.
- Liste soll sortierbar und filterbar sein.

**FR-009: [MVP] Zyklus-Detailansicht**
- Die Steri-Suite soll detaillierte Informationen zu einzelnen Zyklen anzeigen.
- Details sollen Prozessparameter, Zeitstempel, Ergebnis enthalten.

**FR-010: [MVP] Protokoll-Export**
- Die Steri-Suite soll Zyklusprotokolle als PDF oder CSV exportieren können.

#### Traceability

**FR-011: [Growth] Beladungszuordnung**
- Die Steri-Suite soll Beladungen (Trays/Instrumente) Sterilisationszyklen zuordnen können.
- Zuordnung soll für Traceability dokumentiert werden.

### 4.2 Anforderungen an GO-Schnittstelle

#### Allgemein

**FR-012: [MVP] Lokale HTTP-API**
- Die GO-Schnittstelle soll eine RESTful HTTP-API bereitstellen.
- API soll auf localhost erreichbar sein (Standard-Port konfigurierbar).

**FR-013: [MVP] WebSocket-Live-Events**
- Die GO-Schnittstelle soll WebSocket-Verbindungen für Echtzeit-Events unterstützen.
- Events sollen Zyklus-Status-Updates, Gerätestatus-Änderungen enthalten.

**FR-014: [MVP] SQLite-Integration**
- Die GO-Schnittstelle soll alle Daten lokal in SQLite speichern.
- Datenbank soll Geräte, Zyklen, Audit-Logs enthalten.

**FR-015: [MVP] Logging**
- Die GO-Schnittstelle soll technische Logs und Audit-Logs führen.
- Logs sollen verschiedene Levels unterstützen (INFO, ERROR, DEBUG).
- Audit-Logs sollen unveränderbar sein (Write-Ahead-Log, Hashing).

#### Melag-Adapter

**FR-016: [MVP] Melag-Verbindung initialisieren**
- Die GO-Schnittstelle soll Verbindungen zu Melag-Geräten über MELAnet Box initialisieren können.
- Verbindung soll Erreichbarkeitsprüfung (Ping) durchführen.

**FR-017: [MVP] Melag-Zyklus starten**
- Die GO-Schnittstelle soll Startsignale an Melag-Geräte senden können.
- Start-Befehl soll über MELAnet Box (FTP) oder direkte Verbindung erfolgen.

**FR-018: [MVP] Melag-Status abrufen**
- Die GO-Schnittstelle soll Statusinformationen von Melag-Geräten abrufen können.
- Status soll Phase, Parameter, Restzeit enthalten.
- Polling-Intervall: ≤ 2 Sekunden.

**FR-019: [MVP] Melag-Zyklusergebnis abrufen**
- Die GO-Schnittstelle soll abgeschlossene Zyklus-Ergebnisse von Melag-Geräten abrufen können.
- Ergebnisse sollen OK/NOK, Programmart, Timestamp, Fehlercodes enthalten.

#### Getinge-Adapter (Minimalversion)

**FR-020: [MVP] Getinge-ICMP-Monitoring**
- Die GO-Schnittstelle soll ICMP-Ping für Getinge-Geräte durchführen können.
- Ping-Intervall: 10-30 Sekunden (konfigurierbar).
- Ergebnisse sollen in SQLite gespeichert werden (Timestamp, Reachable 0/1).

**FR-021: [Growth] Getinge-Ausfallbenachrichtigung**
- Die GO-Schnittstelle soll bei dauerhaftem Getinge-Ausfall optional Benachrichtigungen senden können.

#### Authentication & Security

**FR-022: [MVP] API-Authentifizierung**
- Die GO-Schnittstelle soll API-Authentifizierung unterstützen.
- Authentifizierung soll über API-Key oder Token erfolgen (konfigurierbar).
- Standard: Optional für lokale Umgebung, erforderlich für Netzwerk-Zugriff.

**FR-023: [MVP] Localhost-Zugriff**
- Die GO-Schnittstelle soll standardmäßig nur über localhost erreichbar sein.
- Netzwerk-Zugriff soll konfigurierbar sein (für Remote-Steri-Suite).

**FR-024: [Growth] Rollenbasierte Zugriffskontrolle**
- Die GO-Schnittstelle soll rollenbasierte Zugriffskontrolle unterstützen können.
- Rollen: Operator, Techniker, QA, Administrator.

#### Testoberfläche

**FR-025: [MVP] Test UI - Device Management**
- Die GO-Schnittstelle soll eine Web-basierte Testoberfläche bereitstellen, die eine Liste aller konfigurierten Geräte anzeigt.
- Die Testoberfläche soll den Verbindungsstatus jedes Geräts anzeigen.
- Die Testoberfläche soll manuelles Verbinden/Trennen von Geräten ermöglichen.
- Die Testoberfläche soll Gerätekonfigurationen anzeigen und bearbeiten ermöglichen.

**FR-026: [MVP] Test UI - API Endpoint Testing**
- Die Testoberfläche soll alle REST-API-Endpunkte der GO-Schnittstelle testbar machen.
- Die Testoberfläche soll Request/Response-Details anzeigen.
- Die Testoberfläche soll HTTP-Status-Codes anzeigen.
- Die Testoberfläche soll WebSocket-Verbindungen testen können.

**FR-027: [MVP] Test UI - Cycle Control Testing**
- Die Testoberfläche soll Test-Zyklen für Melag-Geräte starten können.
- Die Testoberfläche soll Zyklus-Fortschritt in Echtzeit anzeigen.
- Die Testoberfläche soll Zyklus-Status und Parameter anzeigen.
- Die Testoberfläche soll Zyklus-Ergebnisse abrufen können.

**FR-028: [MVP] Test UI - Database Inspection**
- Die Testoberfläche soll gespeicherte Zyklen anzeigen.
- Die Testoberfläche soll Gerätedatensätze inspizieren können.
- Die Testoberfläche soll Audit-Logs abfragen können.
- Die Testoberfläche soll Datenexport (CSV, JSON) ermöglichen.

**FR-029: [MVP] Test UI - Logging and Diagnostics**
- Die Testoberfläche soll Anwendungs-Logs anzeigen.
- Die Testoberfläche soll Logs nach Level filtern können (INFO, ERROR, DEBUG).
- Die Testoberfläche soll Log-Suche ermöglichen.
- Die Testoberfläche soll Log-Export ermöglichen.

**FR-030: [MVP] Test UI - System Status**
- Die Testoberfläche soll System-Gesundheitsstatus anzeigen.
- Die Testoberfläche soll Geräte-Verbindungsstatus anzeigen.
- Die Testoberfläche soll Datenbank-Status anzeigen.
- Die Testoberfläche soll Service-Uptime anzeigen.

**FR-031: [MVP] Test UI - Access Control**
- Die Testoberfläche soll nur über localhost erreichbar sein.
- Die Testoberfläche soll in Produktion deaktivierbar sein (Konfiguration).
- Die Testoberfläche soll einfache Authentifizierung unterstützen (optional).

#### Health Check Endpoints

**FR-032: [MVP] Health Check API**
- Die GO-Schnittstelle soll einen Health-Check-Endpoint bereitstellen (GET /health).
- Der Health-Check soll System-Status zurückgeben (OK, DEGRADED, ERROR).
- Der Health-Check soll Datenbank-Verbindungsstatus prüfen.
- Der Health-Check soll Geräte-Verbindungsstatus prüfen.

**FR-033: [MVP] System Metrics**
- Die GO-Schnittstelle soll System-Metriken-Endpoint bereitstellen (GET /metrics).
- Metriken sollen Service-Uptime enthalten.
- Metriken sollen Anzahl aktiver Geräteverbindungen enthalten.
- Metriken sollen Anzahl verarbeiteter Zyklen enthalten.

**FR-034: [MVP] Diagnostic Endpoints**
- Die GO-Schnittstelle soll Diagnose-Endpoints für Geräteverbindungen bereitstellen.
- Diagnose-Endpoints sollen Verbindungstests ermöglichen.
- Diagnose-Endpoints sollen Protokoll-Debugging-Informationen liefern.
- Diagnose-Endpoints sollen Fehlerprotokolle abrufbar machen.

---

## 5. Nichtfunktionale Anforderungen

### Performance

- Statusaktualisierung Melag: ≤ 1–2 Sekunden
- UI-Refresh ohne Verzögerungen
- Test-UI Seitenladezeit: < 1 Sekunde

### Sicherheit

- Netzwerkkommunikation verschlüsselt, sofern Gerät dies unterstützt
- Lokale Audit-Trails unveränderbar (Write-Ahead-Log, Hashing)
- Test-UI nur über localhost erreichbar
- Test-UI in Produktion deaktivierbar

### Verfügbarkeit

- System muss ohne Adminrechte startbar sein
- Kein Setup, portable Applikation
- Test-UI optional (kann deaktiviert werden)

### Robustheit

- Fehlertoleranz bei Netzunterbrechung
- Saubere Wiederaufnahme laufender Prozesse
- Test-UI soll Systemfehler nicht beeinträchtigen

---

## 6. Datenmodell (SQLite)

### device

| Feld | Typ | Beschreibung |
|------|-----|--------------|
| id | INTEGER PK | Primärschlüssel |
| name | TEXT | Gerätename |
| model | TEXT | Modellbezeichnung |
| manufacturer | TEXT | Melag / Getinge |
| ip | TEXT | IP-Adresse |
| serial | TEXT | Seriennummer |
| type | TEXT | Steri/RDG |
| created | DATETIME | Erstellungsdatum |

### cycle

| Feld | Typ | Beschreibung |
|------|-----|--------------|
| id | INTEGER PK | |
| device_id | FK | |
| program | TEXT | |
| start_ts | DATETIME | |
| end_ts | DATETIME | |
| result | TEXT (OK/NOK) | |
| error_code | TEXT | |
| error_description | TEXT | |

### rdg_status

| timestamp | reachable (0/1) | |

---

## 7. APIs der GO-Schnittstelle

### 7.1 API-Endpunkte Übersicht

#### Melag Endpoints

- **POST /melag/{id}/start** - Startet Sterilisationszyklus
- **GET  /melag/{id}/status** - Ruft aktuellen Status ab
- **GET  /melag/{id}/last-cycle** - Ruft letzten Zyklus ab

#### Getinge Endpoints

- **GET /getinge/{id}/ping** - Prüft Geräteerreichbarkeit

#### Generische Endpoints

- **GET /devices** - Liste aller Geräte
- **POST /devices** - Neues Gerät hinzufügen
- **PUT /devices/{id}** - Gerät aktualisieren
- **DELETE /devices/{id}** - Gerät löschen
- **GET /cycles** - Liste aller Zyklen
- **GET /cycles/{id}** - Einzelner Zyklus

#### Health & Diagnostics

- **GET /health** - System health check
- **GET /metrics** - System metrics
- **GET /diagnostics/{deviceId}** - Device diagnostics

#### Test UI

- **GET /test-ui** - Test interface (HTML)
- **GET /test-ui/api/test** - API endpoint testing
- **GET /test-ui/logs** - Log viewing
- **GET /test-ui/db** - Database inspection

### 7.2 Request/Response Schemas

#### POST /melag/{id}/start

**Request:**
```json
{
  "program": "string",  // Optional: Programmart
  "parameters": {}      // Optional: Zusätzliche Parameter
}
```

**Response (200 OK):**
```json
{
  "cycle_id": "integer",
  "status": "started",
  "message": "Cycle started successfully"
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Device not available",
  "code": "DEVICE_UNAVAILABLE"
}
```

#### GET /melag/{id}/status

**Response (200 OK):**
```json
{
  "device_id": "integer",
  "status": "running|ready|error",
  "cycle_id": "integer",
  "phase": "string",
  "temperature": "float",
  "pressure": "float",
  "time_remaining": "integer",
  "progress_percent": "integer"
}
```

#### GET /devices

**Response (200 OK):**
```json
{
  "devices": [
    {
      "id": "integer",
      "name": "string",
      "model": "string",
      "manufacturer": "string",
      "ip": "string",
      "serial": "string",
      "type": "Steri|RDG",
      "status": "online|offline|error",
      "last_seen": "datetime"
    }
  ]
}
```

#### POST /devices

**Request:**
```json
{
  "name": "string",
  "model": "string",
  "manufacturer": "Melag|Getinge",
  "ip": "string",
  "serial": "string",
  "type": "Steri|RDG",
  "location": "string"  // Optional
}
```

**Response (201 Created):**
```json
{
  "id": "integer",
  "message": "Device created successfully"
}
```

#### GET /health

**Response (200 OK):**
```json
{
  "status": "OK|DEGRADED|ERROR",
  "database": "connected|disconnected",
  "devices": {
    "total": "integer",
    "online": "integer",
    "offline": "integer"
  },
  "uptime_seconds": "integer"
}
```

#### GET /metrics

**Response (200 OK):**
```json
{
  "uptime_seconds": "integer",
  "active_device_connections": "integer",
  "total_cycles_processed": "integer",
  "cycles_today": "integer",
  "api_requests_total": "integer",
  "api_requests_per_minute": "float"
}
```

### 7.3 Authentication Model

**FR-035: [MVP] API-Key Authentication**
- Die GO-Schnittstelle soll API-Key-Authentifizierung unterstützen.
- API-Key wird über HTTP-Header `X-API-Key` übermittelt.
- API-Key wird in Konfigurationsdatei gespeichert.
- Standard: Optional für localhost, erforderlich für Netzwerk-Zugriff.

**FR-036: [Growth] Token-based Authentication**
- Die GO-Schnittstelle soll Token-basierte Authentifizierung unterstützen können.
- Tokens sollen ablaufbar sein (TTL konfigurierbar).
- Token-Erneuerung soll unterstützt werden.

### 7.4 WebSocket Events

**FR-037: [MVP] WebSocket-Verbindung**
- Die GO-Schnittstelle soll WebSocket-Verbindungen für Echtzeit-Events unterstützen.
- Verbindung über `ws://localhost:{port}/ws`

**Event Types:**

**cycle_status_update:**
```json
{
  "event": "cycle_status_update",
  "device_id": "integer",
  "cycle_id": "integer",
  "status": "running|completed|error",
  "phase": "string",
  "progress_percent": "integer",
  "timestamp": "datetime"
}
```

**device_status_change:**
```json
{
  "event": "device_status_change",
  "device_id": "integer",
  "status": "online|offline|error",
  "timestamp": "datetime"
}
```

---

## 8. Benutzerrollen

- Operator (Startet Zyklen, sieht Status)
- Techniker (Geräteadministration)
- Qualitätsbeauftragter (Protokolle, Audit)
- Entwickler/Administrator (Test-UI, Diagnose) - NEU

---

## 9. Fehlerfälle und Behandlung

### Melag

- Gerät nicht erreichbar → Fehleranzeige im UI
- Start nicht möglich → Log + Fehlermeldung
- Abbruch während Zyklus → NOK-Protokoll

### Getinge

- Ping fail → „Offline"
- Dauerhafter Fail → Alarm optional

### Test-UI

- Test-UI Fehler sollen Hauptfunktionalität nicht beeinträchtigen
- Test-UI soll Fehler-Logs anzeigen können

---

## 10. Roadmap

### Phase 1

- Melag Vollintegration
- Getinge Online/Offline
- Lokale Protokolle
- **Test-UI für Entwicklung** - NEU

### Phase 2

- Traceability-Erweiterung
- API-Export ins QM-System
- Erweiterte Diagnose-Funktionen

### Phase 3

- Vollintegration Getinge (abhängig vom Hersteller)

---

## 11. References (NEU)

### Source Documents

- **Research Report:** `docs/research-technical-device-interfaces-2025-11-21.md`
  - MELAnet Box integration approach
  - Device abstraction layer pattern
  - Manufacturer contact requirements

- **Product Brief:** `docs/product-brief-Steri-Connect-Melag-Getinge-GO-2025-11-21.md`
  - Strategic vision
  - MVP scope definition
  - Success metrics

- **Brainstorming Session:** `docs/bmm-brainstorming-session.md`
  - First principles analysis
  - Architecture understanding
  - Core requirements identification

### Technical References

- MELAG Downloadcenter: https://www.melag.com/service/downloadcenter
- Getinge Product Documentation: https://www.getinge.com/de/produkte/
- Go Language Documentation: https://go.dev/doc/

---

---

## 12. MVP Scope Summary

### Phase 1 (MVP) - Functional Requirements

**Steri-Suite Integration:**
- FR-001 bis FR-010: Geräteverwaltung, Zyklussteuerung, Protokollverwaltung

**GO-Schnittstelle Core:**
- FR-012 bis FR-015: HTTP-API, WebSocket, SQLite, Logging

**Melag Integration:**
- FR-016 bis FR-019: Vollständige Melag-Integration

**Getinge Integration:**
- FR-020: ICMP-Monitoring (minimal)

**Test & Diagnostics:**
- FR-025 bis FR-031: Test-UI für Entwicklung
- FR-032 bis FR-034: Health Checks und Diagnostics

**Security:**
- FR-022, FR-023: API-Authentifizierung, Localhost-Zugriff

**Total MVP Requirements:** 34 Functional Requirements

### Phase 2 (Growth) - Functional Requirements

- FR-011: Beladungszuordnung (Traceability)
- FR-021: Getinge-Ausfallbenachrichtigung
- FR-024: Rollenbasierte Zugriffskontrolle
- FR-036: Token-based Authentication

### Phase 3 (Vision) - Future

- Vollständige Getinge-Integration (abhängig von Herstellerfreigabe)
- Erweiterte Traceability-Features
- QM-System-Integration

---

**Document Status:** Complete with numbered FRs (FR-001 to FR-037), API schemas, authentication model, MVP scope marked  
**Total Functional Requirements:** 37  
**MVP Requirements:** 34  
**Next Steps:** Create epics.md with detailed epic and story breakdown  
**Validation:** See `docs/validation-report-prd-2025-11-21.md` for detailed validation results

