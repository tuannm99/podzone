# frontend-v2 (Angular evaluation)

This is a spike, not a replacement yet. `frontend/` (SolidJS) remains the
real app — nothing there has been touched. This project exists to evaluate
Angular as a possible successor, scaffolded with
[Angular CLI](https://github.com/angular/angular-cli) 22.0.6.

## Microfrontend status

Configured as a **native-federation remote** (`@angular-architects/native-federation`),
port `3004`, exposing `./Component` (see `federation.config.mjs`).

This does **not** currently plug into the running `frontend/` host. The host
uses `@originjs/vite-plugin-federation`, an older Webpack-MF-v1-style Vite
plugin; native-federation speaks a different manifest/runtime protocol
(`remoteEntry.json`, `$version: "v4"`, confirmed by curling a live `ng serve`
run). Real cross-framework loading would require migrating the host to
`@module-federation/vite` (MF2) first — a separate, higher-risk change to a
working system, not done as part of this spike. Until then, this app only
runs standalone.

## Development server

To start a local development server, run:

```bash
ng serve --port 3004
```

Once the server is running, open your browser and navigate to `http://localhost:3004/`. The application will automatically reload whenever you modify any of the source files.

## Code scaffolding

Angular CLI includes powerful code scaffolding tools. To generate a new component, run:

```bash
ng generate component component-name
```

For a complete list of available schematics (such as `components`, `directives`, or `pipes`), run:

```bash
ng generate --help
```

## Building

To build the project run:

```bash
ng build
```

This will compile your project and store the build artifacts in the `dist/` directory. By default, the production build optimizes your application for performance and speed.

## Running unit tests

To execute unit tests with the [Vitest](https://vitest.dev/) test runner, use the following command:

```bash
ng test
```

## Running end-to-end tests

For end-to-end (e2e) testing, run:

```bash
ng e2e
```

Angular CLI does not come with an end-to-end testing framework by default. You can choose one that suits your needs.

## Additional Resources

For more information on using the Angular CLI, including detailed command references, visit the [Angular CLI Overview and Command Reference](https://angular.dev/tools/cli) page.
