package ibapi

// IN identifies the msg type of the buf received from TWS or Gateway
type IN = int64

// OUT identifies the msg type of the buf sended to  TWS or Gateway
type OUT = int64
type FiledType = int64
type Version = int

const (
	INT FiledType = 1
	STR FiledType = 2
	FLT FiledType = 3
)

const (
	mTICK_PRICE                               IN = 1
	mTICK_SIZE                                IN = 2
	mORDER_STATUS                             IN = 3
	mERR_MSG                                  IN = 4
	mOPEN_ORDER                               IN = 5
	mACCT_VALUE                               IN = 6
	mPORTFOLIO_VALUE                          IN = 7
	mACCT_UPDATE_TIME                         IN = 8
	mNEXT_VALID_ID                            IN = 9
	mCONTRACT_DATA                            IN = 10
	mEXECUTION_DATA                           IN = 11
	mMARKET_DEPTH                             IN = 12
	mMARKET_DEPTH_L2                          IN = 13
	mNEWS_BULLETINS                           IN = 14
	mMANAGED_ACCTS                            IN = 15
	mRECEIVE_FA                               IN = 16
	mHISTORICAL_DATA                          IN = 17
	mBOND_CONTRACT_DATA                       IN = 18
	mSCANNER_PARAMETERS                       IN = 19
	mSCANNER_DATA                             IN = 20
	mTICK_OPTION_COMPUTATION                  IN = 21
	mTICK_GENERIC                             IN = 45
	mTICK_STRING                              IN = 46
	mTICK_EFP                                 IN = 47
	mCURRENT_TIME                             IN = 49
	mREAL_TIME_BARS                           IN = 50
	mFUNDAMENTAL_DATA                         IN = 51
	mCONTRACT_DATA_END                        IN = 52
	mOPEN_ORDER_END                           IN = 53
	mACCT_DOWNLOAD_END                        IN = 54
	mEXECUTION_DATA_END                       IN = 55
	mDELTA_NEUTRAL_VALIDATION                 IN = 56
	mTICK_SNAPSHOT_END                        IN = 57
	mMARKET_DATA_TYPE                         IN = 58
	mCOMMISSION_REPORT                        IN = 59
	mPOSITION_DATA                            IN = 61
	mPOSITION_END                             IN = 62
	mACCOUNT_SUMMARY                          IN = 63
	mACCOUNT_SUMMARY_END                      IN = 64
	mVERIFY_MESSAGE_API                       IN = 65
	mVERIFY_COMPLETED                         IN = 66
	mDISPLAY_GROUP_LIST                       IN = 67
	mDISPLAY_GROUP_UPDATED                    IN = 68
	mVERIFY_AND_AUTH_MESSAGE_API              IN = 69
	mVERIFY_AND_AUTH_COMPLETED                IN = 70
	mPOSITION_MULTI                           IN = 71
	mPOSITION_MULTI_END                       IN = 72
	mACCOUNT_UPDATE_MULTI                     IN = 73
	mACCOUNT_UPDATE_MULTI_END                 IN = 74
	mSECURITY_DEFINITION_OPTION_PARAMETER     IN = 75
	mSECURITY_DEFINITION_OPTION_PARAMETER_END IN = 76
	mSOFT_DOLLAR_TIERS                        IN = 77
	mFAMILY_CODES                             IN = 78
	mSYMBOL_SAMPLES                           IN = 79
	mMKT_DEPTH_EXCHANGES                      IN = 80
	mTICK_REQ_PARAMS                          IN = 81
	mSMART_COMPONENTS                         IN = 82
	mNEWS_ARTICLE                             IN = 83
	mTICK_NEWS                                IN = 84
	mNEWS_PROVIDERS                           IN = 85
	mHISTORICAL_NEWS                          IN = 86
	mHISTORICAL_NEWS_END                      IN = 87
	mHEAD_TIMESTAMP                           IN = 88
	mHISTOGRAM_DATA                           IN = 89
	mHISTORICAL_DATA_UPDATE                   IN = 90
	mREROUTE_MKT_DATA_REQ                     IN = 91
	mREROUTE_MKT_DEPTH_REQ                    IN = 92
	mMARKET_RULE                              IN = 93
	mPNL                                      IN = 94
	mPNL_SINGLE                               IN = 95
	mHISTORICAL_TICKS                         IN = 96
	mHISTORICAL_TICKS_BID_ASK                 IN = 97
	mHISTORICAL_TICKS_LAST                    IN = 98
	mTICK_BY_TICK                             IN = 99
	mORDER_BOUND                              IN = 100
	mCOMPLETED_ORDER                          IN = 101
	mCOMPLETED_ORDERS_END                     IN = 102
	mREPLACE_FA_END                           IN = 103
)

const (
	mREQ_MKT_DATA                  OUT = 1
	mCANCEL_MKT_DATA               OUT = 2
	mPLACE_ORDER                   OUT = 3
	mCANCEL_ORDER                  OUT = 4
	mREQ_OPEN_ORDERS               OUT = 5
	mREQ_ACCT_DATA                 OUT = 6
	mREQ_EXECUTIONS                OUT = 7
	mREQ_IDS                       OUT = 8
	mREQ_CONTRACT_DATA             OUT = 9
	mREQ_MKT_DEPTH                 OUT = 10
	mCANCEL_MKT_DEPTH              OUT = 11
	mREQ_NEWS_BULLETINS            OUT = 12
	mCANCEL_NEWS_BULLETINS         OUT = 13
	mSET_SERVER_LOGLEVEL           OUT = 14
	mREQ_AUTO_OPEN_ORDERS          OUT = 15
	mREQ_ALL_OPEN_ORDERS           OUT = 16
	mREQ_MANAGED_ACCTS             OUT = 17
	mREQ_FA                        OUT = 18
	mREPLACE_FA                    OUT = 19
	mREQ_HISTORICAL_DATA           OUT = 20
	mEXERCISE_OPTIONS              OUT = 21
	mREQ_SCANNER_SUBSCRIPTION      OUT = 22
	mCANCEL_SCANNER_SUBSCRIPTION   OUT = 23
	mREQ_SCANNER_PARAMETERS        OUT = 24
	mCANCEL_HISTORICAL_DATA        OUT = 25
	mREQ_CURRENT_TIME              OUT = 49
	mREQ_REAL_TIME_BARS            OUT = 50
	mCANCEL_REAL_TIME_BARS         OUT = 51
	mREQ_FUNDAMENTAL_DATA          OUT = 52
	mCANCEL_FUNDAMENTAL_DATA       OUT = 53
	mREQ_CALC_IMPLIED_VOLAT        OUT = 54
	mREQ_CALC_OPTION_PRICE         OUT = 55
	mCANCEL_CALC_IMPLIED_VOLAT     OUT = 56
	mCANCEL_CALC_OPTION_PRICE      OUT = 57
	mREQ_GLOBAL_CANCEL             OUT = 58
	mREQ_MARKET_DATA_TYPE          OUT = 59
	mREQ_POSITIONS                 OUT = 61
	mREQ_ACCOUNT_SUMMARY           OUT = 62
	mCANCEL_ACCOUNT_SUMMARY        OUT = 63
	mCANCEL_POSITIONS              OUT = 64
	mVERIFY_REQUEST                OUT = 65
	mVERIFY_MESSAGE                OUT = 66
	mQUERY_DISPLAY_GROUPS          OUT = 67
	mSUBSCRIBE_TO_GROUP_EVENTS     OUT = 68
	mUPDATE_DISPLAY_GROUP          OUT = 69
	mUNSUBSCRIBE_FROM_GROUP_EVENTS OUT = 70
	mSTART_API                     OUT = 71
	mVERIFY_AND_AUTH_REQUEST       OUT = 72
	mVERIFY_AND_AUTH_MESSAGE       OUT = 73
	mREQ_POSITIONS_MULTI           OUT = 74
	mCANCEL_POSITIONS_MULTI        OUT = 75
	mREQ_ACCOUNT_UPDATES_MULTI     OUT = 76
	mCANCEL_ACCOUNT_UPDATES_MULTI  OUT = 77
	mREQ_SEC_DEF_OPT_PARAMS        OUT = 78
	mREQ_SOFT_DOLLAR_TIERS         OUT = 79
	mREQ_FAMILY_CODES              OUT = 80
	mREQ_MATCHING_SYMBOLS          OUT = 81
	mREQ_MKT_DEPTH_EXCHANGES       OUT = 82
	mREQ_SMART_COMPONENTS          OUT = 83
	mREQ_NEWS_ARTICLE              OUT = 84
	mREQ_NEWS_PROVIDERS            OUT = 85
	mREQ_HISTORICAL_NEWS           OUT = 86
	mREQ_HEAD_TIMESTAMP            OUT = 87
	mREQ_HISTOGRAM_DATA            OUT = 88
	mCANCEL_HISTOGRAM_DATA         OUT = 89
	mCANCEL_HEAD_TIMESTAMP         OUT = 90
	mREQ_MARKET_RULE               OUT = 91
	mREQ_PNL                       OUT = 92
	mCANCEL_PNL                    OUT = 93
	mREQ_PNL_SINGLE                OUT = 94
	mCANCEL_PNL_SINGLE             OUT = 95
	mREQ_HISTORICAL_TICKS          OUT = 96
	mREQ_TICK_BY_TICK_DATA         OUT = 97
	mCANCEL_TICK_BY_TICK_DATA      OUT = 98
	mREQ_COMPLETED_ORDERS          OUT = 99
)

const (
	// mMIN_SERVER_VER_REAL_TIME_BARS       = 34
	// mMIN_SERVER_VER_SCALE_ORDERS         = 35
	// mMIN_SERVER_VER_SNAPSHOT_MKT_DATA    = 35
	// mMIN_SERVER_VER_SSHORT_COMBO_LEGS    = 35
	// mMIN_SERVER_VER_WHAT_IF_ORDERS       = 36
	// mMIN_SERVER_VER_CONTRACT_CONID       = 37
	mMIN_SERVER_VER_PTA_ORDERS                 Version = 39
	mMIN_SERVER_VER_FUNDAMENTAL_DATA           Version = 40
	mMIN_SERVER_VER_DELTA_NEUTRAL              Version = 40
	mMIN_SERVER_VER_CONTRACT_DATA_CHAIN        Version = 40
	mMIN_SERVER_VER_SCALE_ORDERS2              Version = 40
	mMIN_SERVER_VER_ALGO_ORDERS                Version = 41
	mMIN_SERVER_VER_EXECUTION_DATA_CHAIN       Version = 42
	mMIN_SERVER_VER_NOT_HELD                   Version = 44
	mMIN_SERVER_VER_SEC_ID_TYPE                Version = 45
	mMIN_SERVER_VER_PLACE_ORDER_CONID          Version = 46
	mMIN_SERVER_VER_REQ_MKT_DATA_CONID         Version = 47
	mMIN_SERVER_VER_REQ_CALC_IMPLIED_VOLAT     Version = 49
	mMIN_SERVER_VER_REQ_CALC_OPTION_PRICE      Version = 50
	mMIN_SERVER_VER_SSHORTX_OLD                Version = 51
	mMIN_SERVER_VER_SSHORTX                    Version = 52
	mMIN_SERVER_VER_REQ_GLOBAL_CANCEL          Version = 53
	mMIN_SERVER_VER_HEDGE_ORDERS               Version = 54
	mMIN_SERVER_VER_REQ_MARKET_DATA_TYPE       Version = 55
	mMIN_SERVER_VER_OPT_OUT_SMART_ROUTING      Version = 56
	mMIN_SERVER_VER_SMART_COMBO_ROUTING_PARAMS Version = 57
	mMIN_SERVER_VER_DELTA_NEUTRAL_CONID        Version = 58
	mMIN_SERVER_VER_SCALE_ORDERS3              Version = 60
	mMIN_SERVER_VER_ORDER_COMBO_LEGS_PRICE     Version = 61
	mMIN_SERVER_VER_TRAILING_PERCENT           Version = 62
	mMIN_SERVER_VER_DELTA_NEUTRAL_OPEN_CLOSE   Version = 66
	mMIN_SERVER_VER_POSITIONS                  Version = 67
	mMIN_SERVER_VER_ACCOUNT_SUMMARY            Version = 67
	mMIN_SERVER_VER_TRADING_CLASS              Version = 68
	mMIN_SERVER_VER_SCALE_TABLE                Version = 69
	mMIN_SERVER_VER_LINKING                    Version = 70
	mMIN_SERVER_VER_ALGO_ID                    Version = 71
	mMIN_SERVER_VER_OPTIONAL_CAPABILITIES      Version = 72
	mMIN_SERVER_VER_ORDER_SOLICITED            Version = 73
	mMIN_SERVER_VER_LINKING_AUTH               Version = 74
	mMIN_SERVER_VER_PRIMARYEXCH                Version = 75
	mMIN_SERVER_VER_RANDOMIZE_SIZE_AND_PRICE   Version = 76
	mMIN_SERVER_VER_FRACTIONAL_POSITIONS       Version = 101
	mMIN_SERVER_VER_PEGGED_TO_BENCHMARK        Version = 102
	mMIN_SERVER_VER_MODELS_SUPPORT             Version = 103
	mMIN_SERVER_VER_SEC_DEF_OPT_PARAMS_REQ     Version = 104
	mMIN_SERVER_VER_EXT_OPERATOR               Version = 105
	mMIN_SERVER_VER_SOFT_DOLLAR_TIER           Version = 106
	mMIN_SERVER_VER_REQ_FAMILY_CODES           Version = 107
	mMIN_SERVER_VER_REQ_MATCHING_SYMBOLS       Version = 108
	mMIN_SERVER_VER_PAST_LIMIT                 Version = 109
	mMIN_SERVER_VER_MD_SIZE_MULTIPLIER         Version = 110
	mMIN_SERVER_VER_CASH_QTY                   Version = 111
	mMIN_SERVER_VER_REQ_MKT_DEPTH_EXCHANGES    Version = 112
	mMIN_SERVER_VER_TICK_NEWS                  Version = 113
	mMIN_SERVER_VER_REQ_SMART_COMPONENTS       Version = 114
	mMIN_SERVER_VER_REQ_NEWS_PROVIDERS         Version = 115
	mMIN_SERVER_VER_REQ_NEWS_ARTICLE           Version = 116
	mMIN_SERVER_VER_REQ_HISTORICAL_NEWS        Version = 117
	mMIN_SERVER_VER_REQ_HEAD_TIMESTAMP         Version = 118
	mMIN_SERVER_VER_REQ_HISTOGRAM              Version = 119
	mMIN_SERVER_VER_SERVICE_DATA_TYPE          Version = 120
	mMIN_SERVER_VER_AGG_GROUP                  Version = 121
	mMIN_SERVER_VER_UNDERLYING_INFO            Version = 122
	mMIN_SERVER_VER_CANCEL_HEADTIMESTAMP       Version = 123
	mMIN_SERVER_VER_SYNT_REALTIME_BARS         Version = 124
	mMIN_SERVER_VER_CFD_REROUTE                Version = 125
	mMIN_SERVER_VER_MARKET_RULES               Version = 126
	mMIN_SERVER_VER_PNL                        Version = 127
	mMIN_SERVER_VER_NEWS_QUERY_ORIGINS         Version = 128
	mMIN_SERVER_VER_UNREALIZED_PNL             Version = 129
	mMIN_SERVER_VER_HISTORICAL_TICKS           Version = 130
	mMIN_SERVER_VER_MARKET_CAP_PRICE           Version = 131
	mMIN_SERVER_VER_PRE_OPEN_BID_ASK           Version = 132
	mMIN_SERVER_VER_REAL_EXPIRATION_DATE       Version = 134
	mMIN_SERVER_VER_REALIZED_PNL               Version = 135
	mMIN_SERVER_VER_LAST_LIQUIDITY             Version = 136
	mMIN_SERVER_VER_TICK_BY_TICK               Version = 137
	mMIN_SERVER_VER_DECISION_MAKER             Version = 138
	mMIN_SERVER_VER_MIFID_EXECUTION            Version = 139
	mMIN_SERVER_VER_TICK_BY_TICK_IGNORE_SIZE   Version = 140
	mMIN_SERVER_VER_AUTO_PRICE_FOR_HEDGE       Version = 141
	mMIN_SERVER_VER_WHAT_IF_EXT_FIELDS         Version = 142
	mMIN_SERVER_VER_SCANNER_GENERIC_OPTS       Version = 143
	mMIN_SERVER_VER_API_BIND_ORDER             Version = 144
	mMIN_SERVER_VER_ORDER_CONTAINER            Version = 145
	mMIN_SERVER_VER_SMART_DEPTH                Version = 146
	mMIN_SERVER_VER_REMOVE_NULL_ALL_CASTING    Version = 147
	mMIN_SERVER_VER_D_PEG_ORDERS               Version = 148
	mMIN_SERVER_VER_MKT_DEPTH_PRIM_EXCHANGE    Version = 149
	mMIN_SERVER_VER_COMPLETED_ORDERS           Version = 150
	mMIN_SERVER_VER_PRICE_MGMT_ALGO            Version = 151
	mMIN_SERVER_VER_STOCK_TYPE                 Version = 152
	mMIN_SERVER_VER_ENCODE_MSG_ASCII7          Version = 153
	mMIN_SERVER_VER_SEND_ALL_FAMILY_CODES      Version = 154
	mMIN_SERVER_VER_NO_DEFAULT_OPEN_CLOSE      Version = 155
	mMIN_SERVER_VER_PRICE_BASED_VOLATILITY     Version = 156
	mMIN_SERVER_VER_REPLACE_FA_END             Version = 157

	MIN_CLIENT_VER Version = 100
	MAX_CLIENT_VER Version = mMIN_SERVER_VER_REPLACE_FA_END
)

// tick const
const (
	BID_SIZE = iota
	BID
	ASK
	ASK_SIZE
	LAST
	LAST_SIZE
	HIGH
	LOW
	VOLUME
	CLOSE
	BID_OPTION_COMPUTATION
	ASK_OPTION_COMPUTATION
	LAST_OPTION_COMPUTATION
	MODEL_OPTION
	OPEN
	LOW_13_WEEK
	HIGH_13_WEEK
	LOW_26_WEEK
	HIGH_26_WEEK
	LOW_52_WEEK
	HIGH_52_WEEK
	AVG_VOLUME
	OPEN_INTEREST
	OPTION_HISTORICAL_VOL
	OPTION_IMPLIED_VOL
	OPTION_BID_EXCH
	OPTION_ASK_EXCH
	OPTION_CALL_OPEN_INTEREST
	OPTION_PUT_OPEN_INTEREST
	OPTION_CALL_VOLUME
	OPTION_PUT_VOLUME
	INDEX_FUTURE_PREMIUM
	BID_EXCH
	ASK_EXCH
	AUCTION_VOLUME
	AUCTION_PRICE
	AUCTION_IMBALANCE
	MARK_PRICE
	BID_EFP_COMPUTATION
	ASK_EFP_COMPUTATION
	LAST_EFP_COMPUTATION
	OPEN_EFP_COMPUTATION
	HIGH_EFP_COMPUTATION
	LOW_EFP_COMPUTATION
	CLOSE_EFP_COMPUTATION
	LAST_TIMESTAMP
	SHORTABLE
	FUNDAMENTAL_RATIOS
	RT_VOLUME
	HALTED
	BID_YIELD
	ASK_YIELD
	LAST_YIELD
	CUST_OPTION_COMPUTATION
	TRADE_COUNT
	TRADE_RATE
	VOLUME_RATE
	LAST_RTH_TRADE
	RT_HISTORICAL_VOL
	IB_DIVIDENDS
	BOND_FACTOR_MULTIPLIER
	REGULATORY_IMBALANCE
	NEWS_TICK
	SHORT_TERM_VOLUME_3_MIN
	SHORT_TERM_VOLUME_5_MIN
	SHORT_TERM_VOLUME_10_MIN
	DELAYED_BID
	DELAYED_ASK
	DELAYED_LAST
	DELAYED_BID_SIZE
	DELAYED_ASK_SIZE
	DELAYED_LAST_SIZE
	DELAYED_HIGH
	DELAYED_LOW
	DELAYED_VOLUME
	DELAYED_CLOSE
	DELAYED_OPEN
	RT_TRD_VOLUME
	CREDITMAN_MARK_PRICE
	CREDITMAN_SLOW_MARK_PRICE
	DELAYED_BID_OPTION
	DELAYED_ASK_OPTION
	DELAYED_LAST_OPTION
	DELAYED_MODEL_OPTION
	LAST_EXCH
	LAST_REG_TIME
	FUTURES_OPEN_INTEREST
	AVG_OPT_VOLUME
	DELAYED_LAST_TIMESTAMP
	SHORTABLE_SHARES
	NOT_SET
)

// ConnectionState
const (
	DISCONNECTED = iota
	CONNECTING
	CONNECTED
	REDIRECT
)
