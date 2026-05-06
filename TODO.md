# JKit — TODO

> VSCode Todo Tree compatible. Open Command Palette → `Todo Tree: Focus`

## DEVC — Dev Container ✅
- [x] R-DEVC-01 — Generar `.devcontainer/` completo (7 archivos)
- [x] R-DEVC-02 — Sustituir `{{.ProjectName}}`, `{{.JoomlaImage}}`, `{{.Timezone}}` en templates
- [x] R-DEVC-03 — Servicios: Joomla, MariaDB, phpMyAdmin, Mailpit
- [x] R-DEVC-04 — Xdebug preconfigurado
- [x] R-DEVC-05 — Generar `.env` y `.env.example`
- [x] R-DEVC-06 — Agregar `.env` a `.gitignore`
- [x] R-DEVC-07 — Lista de imágenes curadas desde `images.yaml`
- [x] R-DEVC-08 — Permitir imagen manual (`--image`)
- [x] R-DEVC-09 — Solo imágenes Apache + Debian
- [x] R-DEVC-10 — No sobrescribir sin confirmación
- [x] R-DEVC-11 — No asumir versión fija de Joomla/PHP
- [x] R-DEVC-12 — Extensiones VSCode por defecto
- [ ] R-DEVC-13 — Cachear `images.yaml` localmente con aviso de desactualización

## AGNT — Agentes ✅
- [x] R-AGNT-01 — Instalar `gentle-ai` siempre
- [x] R-AGNT-02 — Presentar lista de agentes disponibles
- [x] R-AGNT-03 — Instalar solo agentes elegidos
- [x] R-AGNT-04 — Depositar skill `prd-creator-joomla`
- [x] R-AGNT-05 — Plantillas bash por agente (`templates/agents/*.sh`) con `go:embed`
- [x] R-AGNT-06 — Generar `post-create.sh` dinámicamente
- [x] R-AGNT-07 — No instalar agentes no seleccionados
- [x] R-AGNT-08 — No hardcodear rutas de agentes
- [x] R-AGNT-09 — `jkit agents add/remove`

## EXTG — Extension Generator ✅
- [x] R-EXTG-01 — Soporte: component, module, plugin, template, library, package
- [x] R-EXTG-02 — Ruta correcta en estructura Joomla
- [x] R-EXTG-03 — Convenciones de nomenclatura (`com_`, `mod_`, `plg_`, etc.)
- [x] R-EXTG-04 — Namespaces PSR-4 (Joomla 5/6)
- [x] R-EXTG-05 — Múltiples extensiones en mismo proyecto
- [x] R-EXTG-06 — `jkit build [nombre]` → `.zip` en `builds/`
- [ ] R-EXTG-07 — Agrupar extensiones en único `.zip` (tipo `package`)
- [x] R-EXTG-08 — No generar código Joomla 3
- [x] R-EXTG-09 — No sobrescribir sin confirmación
- [x] R-EXTG-10 — Invocable por usuario, IA vía chat, o CLI
- [x] R-EXTG-11 — Estructura base de tests (PHPUnit)

## INIT — Init & Scaffold ✅
- [x] R-INIT-01 — Solicitar `JOOMLA_SITE_NAME` obligatoriamente
- [x] R-INIT-02 — Defaults `superdev` / `superpassword`
- [x] R-INIT-03 — Modo interactivo (`jkit init`) y parametrizado (`jkit init --name ...`)
- [x] R-INIT-04 — Auto-detectar `.zip` como quickstart o `--quickstart`
- [x] R-INIT-05 — Crear directorio `builds/`
- [x] R-INIT-06 — Orquestar DEVC → AGNT → EXTG → MCPS en `jkit init`
- [x] R-INIT-07 — No lanzar editores
- [x] R-INIT-08 — No sobrescribir sin confirmación
- [x] R-INIT-09 — No instalar todos los agentes sin preguntar
- [x] R-INIT-10 — `jkit create [type]` funcional

## MCPS — MCP Manager ✅
- [x] R-MCPS-01 — MCP de Playwright por defecto
- [x] R-MCPS-02 — MCP de base de datos (MariaDB/MySQL) por defecto
- [x] R-MCPS-03 — Xdebug en modo `trace`/`profile`
- [x] R-MCPS-04 — Delegar ubicación de configs MCP en `gentle-ai`
- [x] R-MCPS-05 — No hardcodear rutas MCP
- [x] R-MCPS-06 — No instalar MCPs no solicitados
- [x] R-MCPS-08 — `jkit mcp add [nombre]`

## TUI — Interfaz Interactiva ⬜
- [ ] Huh — Integrar `huh` (Charmbracelet) para `jkit init` interactivo

## Build & Release ✅
- [x] Install — Script `curl | bash` → `~/.local/bin/`
- [x] Makefile — `make build`, `make test`, `make install`
- [x] CI — GitHub Actions para build + test + release on push to main
