
  export class Config {
    Name: string;
    EncryptConfig?: number;
    Cryptocurrencies: string;
    CurrencyExchangeProvider: string;
    CurrencyPairFormat: CurrencyPairFormat;
    PortfolioAddresses: PortfolioAddresses;
    SMSGlobal: SMSGlobal;
    Webserver: Webserver;
    Exchanges: Exchange[];

    public isConfigCacheValid() : boolean {
        let dateStored = +new Date(window.localStorage['configDate']);
        let dateNow = +new Date();
        var dateDifference = Math.abs(dateNow - dateStored)
        var diffMins = Math.floor((dateDifference / 1000) / 60);
        (diffMins)
    
        if(isNaN(new Date(dateStored).getTime()) || diffMins > 15) {
            return false;
        }
        else {
            return true
        }
    }

    public setConfig(data: any) : void {
        var configData = <Config>data;
        this.Cryptocurrencies = configData.Cryptocurrencies
        this.CurrencyExchangeProvider = configData.CurrencyExchangeProvider
        this.Exchanges = configData.Exchanges
        this.CurrencyPairFormat = configData.CurrencyPairFormat
        this.EncryptConfig = configData.EncryptConfig
        this.Exchanges = configData.Exchanges
        this.Name = configData.Name
        this.PortfolioAddresses = configData.PortfolioAddresses
        this.SMSGlobal = configData.SMSGlobal
        this.Webserver = configData.Webserver
        if(configData.Exchanges.length > 0  
          && configData.Exchanges[0].Pairs 
          && configData.Exchanges[0].Pairs.length > 0) {
            console.log('Successfully retrieved well-formed pairs');
          return;
        }
        this.fromArrayToRedux();
        //Rewrite to cache on parsing to redux array
        this.saveToCache();
    }

    public saveToCache() : void {
      window.localStorage['config'] = JSON.stringify(this); 
      window.localStorage['configDate'] = new Date().toString();
    }

    public clearCache() : void {
      window.localStorage['config'] = null;
      window.localStorage['configDate'] = null;
    }

    public fromArrayToRedux() : void {
        for (var i = 0; i < this.Exchanges.length; i++) {
          this.Exchanges[i].Pairs = new Array<CurrencyPairRedux>();
          var avail = this.Exchanges[i].AvailablePairs.split(',');
          var enabled = this.Exchanges[i].EnabledPairs.split(',');
          for (var j = 0; j < avail.length; j++) {
            var currencyPair = new CurrencyPairRedux();
            currencyPair.Name = avail[j]
            currencyPair.ParsedName = this.stripCurrencyCharacters(avail[j]);
            if (enabled.indexOf(avail[j]) > 0) {
              currencyPair.Enabled = true;
            } else {
              currencyPair.Enabled = false;
            }
            this.Exchanges[i].Pairs.push(currencyPair);
          }
        }

      }

    public parseSettings() : void {

    }

    private stripCurrencyCharacters(name:string) :string {
        name = name.replace('_', '');
        name = name.replace('-', '');
        name = name.replace(' ', '');
        name = name.toLocaleUpperCase();
        return name;
      }

    public fromReduxToArray() : void {
        for (var i = 0; i < this.Exchanges.length; i++) {
          // Step 1, iterate over the Pairs
          var enabled = this.Exchanges[i].EnabledPairs.split(',');
          for (var j = 0; j < this.Exchanges[i].Pairs.length; j++) {
            if (this.Exchanges[i].Pairs[j].Enabled) {
              if (enabled.indexOf(this.Exchanges[i].Pairs[j].Name) == -1) {
                // Step 3 if its not in the enabled list, add it
                enabled.push(this.Exchanges[i].Pairs[j].Name);
              } 
            } else {
              if (enabled.indexOf(this.Exchanges[i].Pairs[j].Name) > -1) {
                enabled.splice(enabled.indexOf(this.Exchanges[i].Pairs[j].Name), 1);
              }
            }
          }
          //Step 4 JSONifiy the enabled list and set it to the this.settings.Exchanges[i].EnabledPairs
          this.Exchanges[i].EnabledPairs = enabled.join();
        }
      }
  }

export class CurrencyPairRedux {
    Name: string;
    ParsedName: string;
    Enabled: boolean;
  }
  
  export interface CurrencyPairFormat {
    Uppercase: boolean;
    Delimiter: string;
  }
  
  export interface PortfolioAddresses {
    Addresses?: Wallet[];
  }

  export interface Wallet {
    Address:string;
    CoinType:string;
    Balance:number;
    Description:string

  }
  
  export interface Contact {
    Name: string;
    Number: string;
    Enabled: boolean;
  }
  
  export interface SMSGlobal {
    Enabled: boolean;
    Username: string;
    Password: string;
    Contacts: Contact[];
  }
  
  export interface Webserver {
    Enabled: boolean;
    AdminUsername: string;
    AdminPassword: string;
    ListenAddress: string;
    WebsocketConnectionLimit: number;
    WebsocketAllowInsecureOrigin: boolean;
  }
  
  export interface ConfigCurrencyPairFormat {
    Uppercase: boolean;
    Index: string;
    Delimiter: string;
  }
  
  export interface RequestCurrencyPairFormat {
    Uppercase: boolean;
    Index: string;
    Delimiter: string;
    Separator: string;
  }
  
  export interface Exchange {
    Name: string;
    Enabled: boolean;
    Verbose: boolean;
    Websocket: boolean;
    RESTPollingDelay: number;
    AuthenticatedAPISupport: boolean;
    APIKey: string;
    APISecret: string;
    AvailablePairs: string;
    EnabledPairs: string;
    BaseCurrencies: string;
    AssetTypes: string;
    ConfigCurrencyPairFormat: ConfigCurrencyPairFormat;
    RequestCurrencyPairFormat: RequestCurrencyPairFormat;
    ClientID: string;
    Pairs: CurrencyPairRedux[];
  }
  
  
  
  