## GoCryptoTrader website
A website interface to interact with the main GoCryptoTrader application. It is developed with Angular 4 with support for Electron

## This is still in active development
 You can track ideas, planned features and what's in progresss on this Trello board: [https://trello.com/b/ZAhMhpOy/gocryptotrader](https://trello.com/b/ZAhMhpOy/gocryptotrader).

## Current Features
+ It can run
+ It can be compiled with Electron to run as an executable
+ Websocket support to listen to GoCryptoTrader events
+ Material design
+ Has a semi-working Settings page
+ Has a basic ticker dashboard


 
## Install dependencies with npm :

``` bash
npm install
```

If you want to generate Angular components with Angular-cli , you **MUST** install `@angular/cli` in npm global context.  
Please follow [Angular-cli documentation](https://github.com/angular/angular-cli) if you had installed a previous version of `angular-cli`.

``` bash
npm install -g @angular/cli
```

## To build for development
npm run web:start

Voila! You can use GoCryptoTrader web app in a local development environment with webpack watching!


## To build for production

- Using development variables (environments/index.ts) :  `npm run electron:dev`
- Using production variables (environments/index.prod.ts) :  `npm run electron:prod`

Your built files are in the /dist folder.

## Included Commands

|Command|Description|
|--|--|
|`npm run start:web`| Execute the app in the brower |
|`npm run electron:linux`| Builds your application and creates an app consumable on linux system |
|`npm run electron:windows`| On a Windows OS, builds your application and creates an app consumable in windows 32/64 bit systems |
|`npm run electron:mac`|  On a MAC OS, builds your application and generates a `.app` file of your application that can be run on Ma |


## Execute E2E tests

You can find end-to-end tests in /e2e folder.

You can run tests with the command lines below : 
- **in a terminal window** -> First, start a web server on port 4200 : `npm run start:web`  
- **in another terminal window** -> Then, launch Protractor (E2E framework): `npm run e2e`

# Contributors 
|User|Github|Contribution|
|--|--|--|
|GloriousCode|https://github.com/gloriouscode |Lead front-end|
|Maxime GRIS|https://github.com/maximegris |Angular4 + Electron Base|
|Shazbert|https://github.com/shazbert |Initial designs|




