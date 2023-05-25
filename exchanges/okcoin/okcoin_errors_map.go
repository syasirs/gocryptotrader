package okcoin

import "errors"

// SetErrorDefaults sets the full error default list
func (o *OKCoin) SetErrorDefaults() {
	o.ErrorCodes = map[string]error{
		"1":     errors.New(`operation failed`),
		"2":     errors.New(`bulk operation partially succeeded`),
		"50000": errors.New(`body cannot be empty`),
		"50001": errors.New(`service temporarily unavailable, please try again later`),
		"50002": errors.New(`json data format error`),
		"50004": errors.New(`endpoint request timeout (does not mean that the request was successful or failed, please check the request result)`),
		"50005": errors.New(`api is offline or unavailable`),
		"50006": errors.New(`invalid content_Type, please use "application/json" format`),
		"50007": errors.New(`account blocked`),
		"50008": errors.New(`user does not exist`),
		"50009": errors.New(`account is suspended due to ongoing liquidation`),
		"50010": errors.New(`user id cannot be empty`),
		"50011": errors.New(`requests too frequent`),
		"50012": errors.New(`account status invalid`),
		"50013": errors.New(`system is busy, please try again later`),
		"50026": errors.New(`system error, please try again later`),
		"50027": errors.New(`the account is restricted from trading`),
		"50028": errors.New(`unable to take the order, please reach out to support center for details`),
		"50030": errors.New(`no permission to use this API`),
		"50032": errors.New(`this asset is blocked, allow its trading and try again`),
		"50033": errors.New(`this instrument is blocked, allow its trading and try again`),
		"50035": errors.New(`this endpoint requires that APIKey must be bound to IP`),
		"50036": errors.New(`invalid expTime`),
		"50037": errors.New(`order expired`),
		"50038": errors.New(`this feature is temporarily unavailable in demo trading`),
		"50039": errors.New(`the before parameter is not available for implementing timestamp pagination`),
		"50041": errors.New(`you are not currently on the whitelist, please contact customer service`),
		"50100": errors.New(`aPI frozen, please contact customer service`),
		"50101": errors.New(`aPIKey does not match current environment`),
		"50102": errors.New(`timestamp request expired`),
		"50103": errors.New(`request header "OK-ACCESS-KEY" cannot be empty`),
		"50104": errors.New(`request header "OK-ACCESS-PASSPHRASE" cannot be empty`),
		"50105": errors.New(`request header "OK-ACCESS-PASSPHRASE" incorrect`),
		"50106": errors.New(`request header "OK-ACCESS-SIGN" cannot be empty`),
		"50107": errors.New(`request header "OK-ACCESS-TIMESTAMP" cannot be empty`),
		"50108": errors.New(`exchange ID does not exist`),
		"50109": errors.New(`exchange domain does not exist`),
		"50111": errors.New(`invalid OK-ACCESS-KEY`),
		"50112": errors.New(`invalid OK-ACCESS-TIMESTAMP`),
		"50113": errors.New(`invalid signature`),
		"50114": errors.New(`invalid authorization`),
		"50115": errors.New(`invalid request method`),
		"51001": errors.New(`instrument ID does not exist`),
		"51003": errors.New(`either client order ID or order ID is required`),
		"51005": errors.New(`order amount exceeds the limit`),
		"51006": errors.New(`order price is not within the price limit (max buy price: {0} min sell price: {1})`),
		"51008": errors.New(`order failed. insufficient account balance, and the adjusted equity in USD is less than IMR`),
		"51009": errors.New(`order placement function is blocked by the platform`),
		"51010": errors.New(`operation is not supported under the current account mode`),
		"51011": errors.New(`duplicated order ID`),
		"51012": errors.New(`token does not exist`),
		"51014": errors.New(`index does not exist`),
		"51015": errors.New(`instrument ID does not match instrument type`),
		"51016": errors.New(`duplicated client order ID`),
		"51020": errors.New(`order amount should be greater than the min available amount`),
		"51023": errors.New(`position does not exist`),
		"51024": errors.New(`trading account is blocked`),
		"51025": errors.New(`order count exceeds the limit`),
		"51026": errors.New(`instrument type does not match underlying index`),
		"51030": errors.New(`funding fee is being settled`),
		"51031": errors.New(`this order price is not within the closing price range`),
		"51032": errors.New(`closing all positions at market price`),
		"51033": errors.New(`the total amount per order for this pair has reached the upper limit`),
		"51037": errors.New(`the current account risk status only supports you to place IOC orders that can reduce the risk of your account`),
		"51038": errors.New(`there is already an IOC order under the current risk module that reduces the risk of the account`),
		"51046": errors.New(`the take profit trigger price should be higher than the order price`),
		"51047": errors.New(`the stop loss trigger price should be lower than the order price`),
		"51048": errors.New(`the take profit trigger price should be lower than the order price`),
		"51049": errors.New(`the stop loss trigger price should be higher than the order price`),
		"51050": errors.New(`the take profit trigger price should be higher than the best ask price`),
		"51051": errors.New(`the stop loss trigger price should be lower than the best ask price`),
		"51052": errors.New(`the take profit trigger price should be lower than the best bid price`),
		"51053": errors.New(`the stop loss trigger price should be higher than the best bid price`),
		"51054": errors.New(`getting information timed out, please try again later`),
		"51056": errors.New(`action not allowed`),
		"51058": errors.New(`no available position for this algo order`),
		"51059": errors.New(`strategy for the current state does not support this operation`),
		"51101": errors.New(`entered amount exceeds the max pending order amount (Cont) per transaction`),
		"51103": errors.New(`entered amount exceeds the max pending order count of the underlying asset`),
		"51104": errors.New(`entered amount exceeds the max pending order amount (Cont) of the underlying asset`),
		"51106": errors.New(`entered amount exceeds the max order amount (Cont) of the underlying asset`),
		"51107": errors.New(`entered amount exceeds the max holding amount (Cont)`),
		"51109": errors.New(`no available offer`),
		"51110": errors.New(`you can only place a limit order after call auction has started`),
		"51112": errors.New(`close order size exceeds your available size`),
		"51113": errors.New(`market-price liquidation requests too frequent`),
		"51115": errors.New(`cancel all pending close-orders before liquidation`),
		"51117": errors.New(`pending close-orders count exceeds limit`),
		"51121": errors.New(`order count should be the integer multiples of the lot size`),
		"51124": errors.New(`you can only place limit orders during call auction`),
		"51127": errors.New(`available balance is 0`),
		"51129": errors.New(`the value of the position and buy order has reached the position limit, and no further buying is allowed`),
		"51131": errors.New(`insufficient balance`),
		"51132": errors.New(`your position amount is negative and less than the minimum trading amount`),
		"51134": errors.New(`closing position failed. Please check your holdings and pending orders`),
		"51139": errors.New(`reduce-only feature is unavailable for the spot transactions by simple account`),
		"51143": errors.New(`there is no valid quotation in the market, and the order cannot be filled in USDT mode, please try to switch to currency mode`),
		"51148": errors.New(`reduce-only cannot increase the position quantity`),
		"51149": errors.New(`order timed out, please try again later`),
		"51150": errors.New(`the precision of the number of trades or the price exceeds the limit`),
		"51201": errors.New(`value of per market order cannot exceed 1,000,000 USDT`),
		"51202": errors.New(`market - order amount exceeds the max amount`),
		"51204": errors.New(`the price for the limit order cannot be empty`),
		"51205": errors.New(`reduce-only is not available`),
		"51250": errors.New(`algo order price is out of the available range`),
		"51251": errors.New(`algo order type error (when user place an iceberg order)`),
		"51252": errors.New(`algo order amount is out of the available range`),
		"51253": errors.New(`average amount exceeds the limit of per iceberg order`),
		"51254": errors.New(`iceberg average amount error (when user place an iceberg order)`),
		"51255": errors.New(`limit of per iceberg order: Total amount/1000 < x <= Total amount`),
		"51256": errors.New(`iceberg order price variance error`),
		"51257": errors.New(`trail order callback rate error`),
		"51258": errors.New(`trail - order placement failed. The trigger price of a sell order should be higher than the last transaction price`),
		"51259": errors.New(`trail - order placement failed. The trigger price of a buy order should be lower than the last transaction price`),
		"51264": errors.New(`average amount exceeds the limit of per time-weighted order`),
		"51265": errors.New(`time-weighted order limit error`),
		"51267": errors.New(`time-weighted order strategy initiative rate error`),
		"51268": errors.New(`time-weighted order strategy initiative range error`),
		"51270": errors.New(`the limit of time-weighted order price variance is 0 < x <= 1%`),
		"51271": errors.New(`sweep ratio should be 0 < x <= 100%`),
		"51272": errors.New(`price variance should be 0 < x <= 1%`),
		"51274": errors.New(`total quantity of time-weighted order must be larger than single order limit`),
		"51275": errors.New(`the amount of single stop-market order cannot exceed the upper limit`),
		"51276": errors.New(`stop - Market orders cannot specify a price`),
		"51277": errors.New(`tp trigger price cannot be higher than the last price`),
		"51278": errors.New(`sl trigger price cannot be lower than the last price`),
		"51279": errors.New(`tp trigger price cannot be lower than the last price`),
		"51280": errors.New(`sl trigger price cannot be higher than the last price`),
		"51281": errors.New(`trigger not support the tgtCcy parameter`),
		"51288": errors.New(`we are stopping the Bot. Please do not click it multiple times`),
		"51289": errors.New(`bot configuration does not exist. Please try again later`),
		"51290": errors.New(`the Bot engine is being upgraded. Please try again later`),
		"51291": errors.New(`this Bot does not exist or has been stopped`),
		"51292": errors.New(`this Bot type does not exist`),
		"51293": errors.New(`this Bot does not exist`),
		"51294": errors.New(`this Bot cannot be created temporarily. Please try again later`),
		"51300": errors.New(`tp trigger price cannot be higher than the mark price`),
		"51302": errors.New(`sl trigger price cannot be lower than the mark price`),
		"51303": errors.New(`tp trigger price cannot be lower than the mark price`),
		"51304": errors.New(`sl trigger price cannot be higher than the mark price`),
		"51305": errors.New(`tp trigger price cannot be higher than the index price`),
		"51306": errors.New(`sl trigger price cannot be lower than the index price`),
		"51307": errors.New(`tp trigger price cannot be lower than the index price`),
		"51308": errors.New(`sl trigger price cannot be higher than the index price`),
		"51309": errors.New(`cannot create trading bot during call auction`),
		"51313": errors.New(`manual transfer in isolated mode does not support bot trading`),
		"51341": errors.New(`position closing not allowed`),
		"51342": errors.New(`closing order already exists. Please try again later`),
		"51343": errors.New(`tp price must be less than the lower price`),
		"51344": errors.New(`sl price must be greater than the upper price`),
		"51345": errors.New(`policy type is not grid policy`),
		"51346": errors.New(`the highest price cannot be lower than the lowest price`),
		"51347": errors.New(`no profit available`),
		"51348": errors.New(`stop loss price should be less than the lower price in the range`),
		"51349": errors.New(`stop profit price should be greater than the highest price in the range`),
		"51350": errors.New(`no recommended parameters`),
		"51351": errors.New(`single income must be greater than 0`),
		"51400": errors.New(`cancelation failed as the order does not exist`),
		"51401": errors.New(`cancelation failed as the order is already canceled`),
		"51402": errors.New(`cancelation failed as the order is already completed`),
		"51403": errors.New(`cancelation failed as the order type does not support cancelation`),
		"51404": errors.New(`order cancellation unavailable during the second phase of call auction`),
		"51405": errors.New(`cancelation failed as you do not have any pending orders`),
		"51407": errors.New(`either order ID or client order ID is required`),
		"51408": errors.New(`pair id or name does not match the order info`),
		"51409": errors.New(`either pair id or pair name id is required`),
		"51410": errors.New(`cancelation pending. duplicate order rejected`),
		"51411": errors.New(`account does not have permission for mass cancelation`),
		"51412": errors.New(`the order has been triggered and cannot be canceled`),
		"51413": errors.New(`cancelation failed as the order type is not supported by endpoint`),
		"51415": errors.New(`unable to place order. spot trading only supports using the last price as trigger price. please select "Last" and try again`),
		"51500": errors.New(`either order price or amount is required`),
		"51503": errors.New(`order modification failed as the order does not exist`),
		"51506": errors.New(`order modification unavailable for the order type`),
		"51508": errors.New(`orders are not allowed to be modified during the call auction`),
		"51509": errors.New(`modification failed as the order has been canceled`),
		"51510": errors.New(`modification failed as the order has been completed`),
		"51511": errors.New(`operation failed as the order price did not meet the requirement for post only`),
		"51512": errors.New(`failed to amend orders in batches. you cannot have duplicate orders in the same amend-batch-orders request`),
		"51513": errors.New(`number of modification requests that are currently in progress for an order cannot exceed 3`),
		"51600": errors.New(`status not found`),
		"51601": errors.New(`order status and order ID cannot exist at the same time`),
		"51602": errors.New(`either order status or order ID is required`),
		"51603": errors.New(`order does not exist`),
		"51607": errors.New(`the file is generating`),
		"52000": errors.New(`no market data found`),
		"54000": errors.New(`margin trading is not supported`),
		"58002": errors.New(`please activate Savings Account first`),
		"58003": errors.New(`currency type is not supported by Savings Account`),
		"58004": errors.New(`account blocked`),
		"58007": errors.New(`abnormal Assets interface. Please try again later`),
		"58008": errors.New(`you do not have assets in this currency`),
		"58009": errors.New(`currency pair do not exist`),
		"58100": errors.New(`the trading product triggers risk control, and the platform has suspended the fund transfer-out function with related users. Please wait patiently`),
		"58101": errors.New(`transfer suspended`),
		"58102": errors.New(`too frequent transfer (transfer too frequently)`),
		"58104": errors.New(`since your P2P transaction is abnormal, you are restricted from making fund transfers. Please contact customer support to remove the restriction`),
		"58105": errors.New(`since your P2P transaction is abnormal, you are restricted from making fund transfers. Please transfer funds on our website or app to complete identity verification`),
		"58112": errors.New(`your fund transfer failed. Please try again later`),
		"58114": errors.New(`transfer amount must be more than 0`),
		"58115": errors.New(`sub-account does not exist`),
		"58116": errors.New(`transfer amount exceeds the limit`),
		"58117": errors.New(`account assets are abnormal, please deal with negative assets before transferring`),
		"58120": errors.New(`the transfer service is temporarily unavailable, please try again later`),
		"58121": errors.New(`this transfer will result in a high-risk level of your position, which may lead to forced liquidation. You need to re-adjust the transfer amount to make sure the position is at a safe level before proceeding with the transfer`),
		"58123": errors.New(`parameter from cannot equal to parameter to`),
		"58201": errors.New(`withdrawal amount exceeds the daily limit`),
		"58202": errors.New(`the minimum withdrawal amount for NEO is 1, and the amount must be an integer`),
		"58203": errors.New(`please add a withdrawal address`),
		"58204": errors.New(`withdrawal suspended`),
		"58205": errors.New(`withdrawal amount exceeds the upper limit`),
		"58206": errors.New(`withdrawal amount is less than the lower limit`),
		"58207": errors.New(`withdrawal address is not in the verification-free whitelist`),
		"58208": errors.New(`withdrawal failed. Please link your email`),
		"58209": errors.New(`sub-accounts cannot be deposits or withdrawals`),
		"58210": errors.New(`withdrawal fee exceeds the upper limit`),
		"58211": errors.New(`withdrawal fee is lower than the lower limit (withdrawal endpoint: incorrect fee)`),
		"58213": errors.New(`please set a trading password before withdrawing`),
		"58215": errors.New(`withdrawal id does not exist`),
		"58216": errors.New(`operation not allowed`),
		"58217": errors.New(`you cannot withdraw your asset at the moment due to a risk detected in your withdrawal address, contact customer support for details`),
		"58218": errors.New(`your saved withdrawal account has expired`),
		"58220": errors.New(`the withdrawal order is already canceled`),
		"58221": errors.New(`missing label of withdrawal address`),
		"58222": errors.New(`temporarily unable to process withdrawal address`),
		"58224": errors.New(`this type of coin does not support on-chain withdrawals. please use internal transfers`),
		"58300": errors.New(`deposit-address count exceeds the limit`),
		"58301": errors.New(`deposit-address not exist`),
		"58302": errors.New(`deposit-address needs tag`),
		"58304": errors.New(`failed to create invoice`),
		"58350": errors.New(`insufficient balance`),
		"58351": errors.New(`invoice expired`),
		"58352": errors.New(`invalid invoice`),
		"58353": errors.New(`deposit amount must be within limits`),
		"58354": errors.New(`you have reached the limit of 10,000 invoices per day`),
		"58355": errors.New(`permission denied. Please contact your account manager`),
		"58356": errors.New(`the accounts of the same node do not support the Lightning network deposit or withdrawal`),
		"58358": errors.New(`fromCcy should not be the same as toCcy`),
		"58370": errors.New(`the daily usage of small assets convert exceeds the limit`),
		"58371": errors.New(`small assets exceed the maximum limit`),
		"58372": errors.New(`insufficient small assets`),
		"59000": errors.New(`your settings failed as you have positions or open orders`),
		"59002": errors.New(`sub-account settings failed as it has positions, open orders, or trading bots`),
		"59004": errors.New(`only ids with the same instrument type are supported`),
		"59200": errors.New(`insufficient account balance`),
		"59201": errors.New(`negative account balance`),
		"59401": errors.New(`holdings already reached the limit`),
		"59402": errors.New(`none of the passed instId is in live state, please check them separately`),
		"59500": errors.New(`only the APIKey of the main account has permission`),
		"59501": errors.New(`only 50 APIKeys can be created per account`),
		"59502": errors.New(`note name cannot be duplicate with the currently created APIKey note name`),
		"59503": errors.New(`each APIKey can bind up to 20 IP addresses`),
		"59504": errors.New(`the sub account does not support the withdrawal function`),
		"59505": errors.New(`the passphrase format is incorrect`),
		"59506": errors.New(`aPIKey does not exist`),
		"59507": errors.New(`the two accounts involved in a transfer must be two different sub accounts under the same parent account`),
		"59510": errors.New(`sub-account does not exist`),
		"59601": errors.New(`this sub-account name already exists, try another name`),
		"59602": errors.New(`number of api keys exceeds the limit`),
		"59603": errors.New(`number of sub accounts exceeds the limit`),
		"59604": errors.New(`only the main account APIkey can access this api`),
		"59605": errors.New(`this API key does not exist in your sub-account, try another API key`),
		"59606": errors.New(`transfer funds to your main account before deleting your sub-account`),
		"59612": errors.New(`cannot convert time format`),
		"59613": errors.New(`there is currently no escrow relationship established with the sub account`),
		"59614": errors.New(`managed sub account do not support this operation`),
		"59615": errors.New(`the time interval between the begin date and end date cannot exceed 180 days`),
		"59616": errors.New(`begin date cannot be greater than end date`),
		"59617": errors.New(`sub-account created. failed to set up account level`),
		"59618": errors.New(`failed to create sub-account`),
	}
}

var websocketErrorCodes = map[string]string{
	"1":     "Operation failed.",
	"2":     "Bulk operation partially succeeded.",
	"50000": "Body cannot be empty.",
	"50001": "Service temporarily unavailable, please try again later.",
	"50002": "Json data format error.",
	"50004": "Endpoint request timeout (does not mean that the request was successful or failed, please check the request result).",
	"50005": "API is offline or unavailable.",
	"50006": "Invalid Content_Type, please use 'application/json' format.",
	"50007": "Account blocked.",
	"50008": "User does not exist.",
	"50009": "Account is suspended due to ongoing liquidation.",
	"50010": "User ID cannot be empty.",
	"50011": "Requests too frequent.",
	"50012": "Account status invalid.",
	"50013": "System is busy, please try again later.",
	"50026": "System error, please try again later.",
	"50027": "The account is restricted from trading.",
	"50028": "Unable to take the order, please reach out to support center for details.",
	"50030": "No permission to use this API",
	"50032": "This asset is blocked, allow its trading and try again",
	"50033": "This instrument is blocked, allow its trading and try again",
	"50035": "This endpoint requires that APIKey must be bound to IP",
	"50036": "Invalid expTime",
	"50037": "Order expired",
	"50038": "This feature is temporarily unavailable in demo trading",
	"50039": "The before parameter is not available for implementing timestamp pagination",
	"50041": "You are not currently on the whitelist, please contact customer service",
	"50100": `API frozen, please contact customer service`,
	"50101": `APIKey does not match current environment`,
	"50102": `Timestamp request expired`,
	"50103": `Request header "OK-ACCESS-KEY" cannot be empty`,
	"50104": `Request header "OK-ACCESS-PASSPHRASE" cannot be empty`,
	"50105": `Request header "OK-ACCESS-PASSPHRASE" incorrect`,
	"50106": `Request header "OK-ACCESS-SIGN" cannot be empty`,
	"50107": `Request header "OK-ACCESS-TIMESTAMP" cannot be empty`,
	"50108": `Exchange ID does not exist`,
	"50109": `Exchange domain does not exist`,
	"50111": `Invalid OK-ACCESS-KEY`,
	"50112": `Invalid OK-ACCESS-TIMESTAMP`,
	"50113": `Invalid signature`,
	"50114": `Invalid authorization`,
	"50115": `Invalid request method`,
	"51001": `Instrument ID does not exist`,
	"51003": `Either client order ID or order ID is required`,
	"51005": `Order amount exceeds the limit`,
	"51009": `Order placement function is blocked by the platform`,
	"51010": `Operation is not supported under the current account mode`,
	"51011": `Duplicated order ID`,
	"51012": `Token does not exist`,
	"51014": `Index does not exist`,
	"51015": `Instrument ID does not match instrument type`,
	"51016": `Duplicated client order ID`,
	"51020": `Order amount should be greater than the min available amount`,
	"51023": `Position does not exist`,
	"51024": `Trading account is blocked`,
	"51025": `Order count exceeds the limit`,
	"51026": `Instrument type does not match underlying index`,
	"51030": `Funding fee is being settled`,
	"51031": `This order price is not within the closing price range`,
	"51032": `Closing all positions at market price`,
	"51033": `The total amount per order for this pair has reached the upper limit`,
	"51037": `The current account risk status only supports you to place IOC orders that can reduce the risk of your account`,
	"51038": `There is already an IOC order under the current risk module that reduces the risk of the account`,
	"51046": `The take profit trigger price should be higher than the order price`,
	"51047": `The stop loss trigger price should be lower than the order price`,
	"51048": `The take profit trigger price should be lower than the order price`,
	"51049": `The stop loss trigger price should be higher than the order price`,
	"51050": `The take profit trigger price should be higher than the best ask price`,
	"51051": `The stop loss trigger price should be lower than the best ask price`,
	"51052": `The take profit trigger price should be lower than the best bid price`,
	"51053": `The stop loss trigger price should be higher than the best bid price`,
	"51054": `Getting information timed out, please try again later`,
	"51056": `Action not allowed`,
	"51058": `No available position for this algo order`,
	"51059": `Strategy for the current state does not support this operation`,
	"51101": `Entered amount exceeds the max pending order amount (Cont) per transaction`,
	"51103": `Entered amount exceeds the max pending order count of the underlying asset`,
	"51104": `Entered amount exceeds the max pending order amount (Cont) of the underlying asset`,
	"51106": `Entered amount exceeds the max order amount (Cont) of the underlying asset`,
	"51107": `Entered amount exceeds the max holding amount (Cont)`,
	"51109": `No available offer`,
	"51110": `You can only place a limit order after Call Auction has started`,
	"51112": `Close order size exceeds your available size`,
	"51113": `Market-price liquidation requests too frequent`,
	"51115": `Cancel all pending close-orders before liquidation`,
	"51117": `Pending close-orders count exceeds limit`,
	"51121": `Order count should be the integer multiples of the lot size`,
	"51124": `You can only place limit orders during call auction`,
	"51127": `Available balance is 0`,
	"51129": `The value of the position and buy order has reached the position limit, and no further buying is allowed`,
	"51131": `Insufficient balance`,
	"51132": `Your position amount is negative and less than the minimum trading amount`,
	"51134": `Closing position failed. Please check your holdings and pending orders`,
	"51139": `Reduce-only feature is unavailable for the spot transactions by simple account`,
	"51143": `There is no valid quotation in the market, and the order cannot be filled in USDT mode, please try to switch to currency mode`,
	"51148": `ReduceOnly cannot increase the position quantity`,
	"51149": `Order timed out, please try again later`,
	"51150": `The precision of the number of trades or the price exceeds the limit`,
	"51201": `Value of per market order cannot exceed 1,000,000 USDT`,
	"51202": `Market - order amount exceeds the max amount`,
	"51204": `The price for the limit order cannot be empty`,
	"51205": `Reduce-Only is not available`,
	"51250": `Algo order price is out of the available range`,
	"51251": `Algo order type error (when user place an iceberg order)`,
	"51252": `Algo order amount is out of the available range`,
	"51253": `Average amount exceeds the limit of per iceberg order`,
	"51254": `Iceberg average amount error (when user place an iceberg order)`,
	"51255": `Limit of per iceberg order: Total amount/1000 < x <= Total amount`,
	"51256": `Iceberg order price variance error`,
	"51257": `Trail order callback rate error`,
	"51258": `Trail - order placement failed. The trigger price of a sell order should be higher than the last transaction price`,
	"51259": `Trail - order placement failed. The trigger price of a buy order should be lower than the last transaction price`,
	"51264": `Average amount exceeds the limit of per time-weighted order`,
	"51265": `Time-weighted order limit error`,
	"51267": `Time-weighted order strategy initiative rate error`,
	"51268": `Time-weighted order strategy initiative range error`,
	"51270": `The limit of time-weighted order price variance is 0 < x <= 1%`,
	"51271": `Sweep ratio should be 0 < x <= 100%`,
	"51272": `Price variance should be 0 < x <= 1%`,
	"51274": `Total quantity of time-weighted order must be larger than single order limit`,
	"51275": `The amount of single stop-market order cannot exceed the upper limit`,
	"51276": `Stop - Market orders cannot specify a price`,
	"51277": `TP trigger price cannot be higher than the last price`,
	"51278": `SL trigger price cannot be lower than the last price`,
	"51279": `TP trigger price cannot be lower than the last price`,
	"51280": `SL trigger price cannot be higher than the last price`,
	"51281": `trigger not support the tgtCcy parameter`,
	"51288": `We are stopping the Bot. Please do not click it multiple times`,
	"51289": `Bot configuration does not exist. Please try again later`,
	"51290": `The Bot engine is being upgraded. Please try again later`,
	"51291": `This Bot does not exist or has been stopped`,
	"51292": `This Bot type does not exist`,
	"51293": `This Bot does not exist`,
	"51294": `This Bot cannot be created temporarily. Please try again later`,
	"51300": `TP trigger price cannot be higher than the mark price`,
	"51302": `SL trigger price cannot be lower than the mark price`,
	"51303": `TP trigger price cannot be lower than the mark price`,
	"51304": `SL trigger price cannot be higher than the mark price`,
	"51305": `TP trigger price cannot be higher than the index price`,
	"51306": `SL trigger price cannot be lower than the index price`,
	"51307": `TP trigger price cannot be lower than the index price`,
	"51308": `SL trigger price cannot be higher than the index price`,
	"51309": `Cannot create trading bot during call auction`,
	"51313": `Manual transfer in isolated mode does not support bot trading`,
	"51341": `Position closing not allowed`,
	"51342": `Closing order already exists. Please try again later`,
	"51343": `TP price must be less than the lower price`,
	"51344": `SL price must be greater than the upper price`,
	"51345": `Policy type is not grid policy`,
	"51346": `The highest price cannot be lower than the lowest price`,
	"51347": `No profit available`,
	"51348": `Stop loss price should be less than the lower price in the range`,
	"51349": `Stop profit price should be greater than the highest price in the range`,
	"51350": `No recommended parameters`,
	"51351": `Single income must be greater than 0`,
	"51400": `cancellation failed as the order does not exist`,
	"51401": `cancellation failed as the order is already canceled`,
	"51402": `cancellation failed as the order is already completed`,
	"51403": `cancellation failed as the order type does not support cancellation`,
	"51404": `Order cancellation unavailable during the second phase of call auction`,
	"51405": `cancellation failed as you do not have any pending orders`,
	"51407": `Either order ID or client order ID is required`,
	"51408": `Pair ID or name does not match the order info`,
	"51409": `Either pair ID or pair name ID is required`,
	"51410": `cancellation pending. Duplicate order rejected`,
	"51411": `Account does not have permission for mass cancellation`,
	"51412": `The order has been triggered and cannot be canceled`,
	"51413": `cancellation failed as the order type is not supported by endpoint`,
	"51415": `Unable to place order. Spot trading only supports using the last price as trigger price. Please select "Last" and try again`,
	"51500": `Either order price or amount is required`,
	"51503": `Order modification failed as the order does not exist`,
	"51506": `Order modification unavailable for the order type`,
	"51508": `Orders are not allowed to be modified during the call auction`,
	"51509": `Modification failed as the order has been canceled`,
	"51510": `Modification failed as the order has been completed`,
	"51511": `Operation failed as the order price did not meet the requirement for Post Only`,
	"51512": `Failed to amend orders in batches. You cannot have duplicate orders in the same amend-batch-orders request`,
	"51513": `Number of modification requests that are currently in progress for an order cannot exceed 3`,
	"51600": `Status not found`,
	"51601": `Order status and order ID cannot exist at the same time`,
	"51602": `Either order status or order ID is required`,
	"51603": `Order does not exist`,
	"51607": `The file is generating`,
}
