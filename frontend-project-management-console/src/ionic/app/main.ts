import { platformBrowserDynamic } from "@angular/platform-browser-dynamic";

import { AppModule } from "./app.module";
import { enableProdMode } from "@angular/core";

import { KeycloakService } from "../services/keycloak/keycloak.service";

enableProdMode();
platformBrowserDynamic().bootstrapModule(AppModule);

// KeycloakService.init({ onLoad: 'check-sso', checkLoginIframeInterval: 1 })
//     .then(() => {
//         platformBrowserDynamic().bootstrapModule(AppModule);
// })
// .catch((e: string) => {
//     console.log('Error in ng4 bootstrap: ' + e);
// });
