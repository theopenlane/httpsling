package httpsling

const (
	// Authentication
	HeaderAuthorization      = "Authorization"
	HeaderProxyAuthenticate  = "Proxy-Authenticate"
	HeaderProxyAuthorization = "Proxy-Authorization"
	HeaderWWWAuthenticate    = "WWW-Authenticate"

	BearerAuthHeader = "Bearer "
	BasicAuthHeader  = "Basic "

	// Caching
	HeaderAge           = "Age"
	HeaderCacheControl  = "Cache-Control"
	HeaderClearSiteData = "Clear-Site-Data"
	HeaderExpires       = "Expires"
	HeaderPragma        = "Pragma"
	HeaderWarning       = "Warning"

	// Client hints
	HeaderAcceptCH         = "Accept-CH"
	HeaderAcceptCHLifetime = "Accept-CH-Lifetime"
	HeaderContentDPR       = "Content-DPR"
	HeaderDPR              = "DPR"
	HeaderEarlyData        = "Early-Data"
	HeaderSaveData         = "Save-Data"
	HeaderViewportWidth    = "Viewport-Width"
	HeaderWidth            = "Width"

	// Conditionals
	HeaderETag              = "ETag"
	HeaderIfMatch           = "If-Match"
	HeaderIfModifiedSince   = "If-Modified-Since"
	HeaderIfNoneMatch       = "If-None-Match"
	HeaderIfUnmodifiedSince = "If-Unmodified-Since"
	HeaderLastModified      = "Last-Modified"
	HeaderVary              = "Vary"

	// Connection management
	HeaderConnection      = "Connection"
	HeaderKeepAlive       = "Keep-Alive"
	HeaderProxyConnection = "Proxy-Connection"

	// Content negotiation
	HeaderAccept         = "Accept"
	HeaderAcceptCharset  = "Accept-Charset"
	HeaderAcceptEncoding = "Accept-Encoding"
	HeaderAcceptLanguage = "Accept-Language"

	// Controls
	HeaderCookie      = "Cookie"
	HeaderExpect      = "Expect"
	HeaderMaxForwards = "Max-Forwards"
	HeaderSetCookie   = "Set-Cookie"

	// CORS
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderOrigin                        = "Origin"
	HeaderTimingAllowOrigin             = "Timing-Allow-Origin"
	HeaderXPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies"

	// Do Not Track
	HeaderDNT = "DNT"
	HeaderTk  = "Tk"

	// Downloads
	HeaderContentDisposition = "Content-Disposition"

	// Message body information
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLanguage = "Content-Language"
	HeaderContentLength   = "Content-Length"
	HeaderContentLocation = "Content-Location"
	HeaderContentType     = "Content-Type"

	// Content Types
	ContentTypeForm                   = "application/x-www-form-urlencoded" // https://datatracker.ietf.org/doc/html/rfc1866
	ContentTypeMultipart              = "multipart/form-data"               // https://datatracker.ietf.org/doc/html/rfc2388
	ContentTypeJSON                   = "application/json"                  // https://datatracker.ietf.org/doc/html/rfc4627
	ContentTypeJSONUTF8               = "application/json;charset=utf-8"    // https://datatracker.ietf.org/doc/html/rfc4627
	ContentTypeXML                    = "application/xml"                   // https://datatracker.ietf.org/doc/html/rfc3023
	ContentTypeXMLUTF8                = "application/xml;charset=utf-8"
	ContentTypeYAML                   = "application/yaml" // https://www.rfc-editor.org/rfc/rfc9512.html
	ContentTypeYAMLUTF8               = "application/yaml;charset=utf-8"
	ContentTypeText                   = "text/plain"
	ContentTypeTextUTF8               = "text/plain;charset=utf-8"
	ContentTypeApplicationOctetStream = "application/octet-stream"

	// Proxies
	HeaderForwarded       = "Forwarded"
	HeaderVia             = "Via"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedHost  = "X-Forwarded-Host"
	HeaderXForwardedProto = "X-Forwarded-Proto"

	// Redirects
	HeaderLocation = "Location"

	// Request context
	HeaderFrom           = "From"
	HeaderHost           = "Host"
	HeaderReferer        = "Referer"
	HeaderReferrerPolicy = "Referrer-Policy"
	HeaderUserAgent      = "User-Agent"

	// Response context
	HeaderAllow  = "Allow"
	HeaderServer = "Server"

	// Range requests.
	HeaderAcceptRanges = "Accept-Ranges"
	HeaderContentRange = "Content-Range"
	HeaderIfRange      = "If-Range"
	HeaderRange        = "Range"

	// Security
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderCrossOriginResourcePolicy       = "Cross-Origin-Resource-Policy"
	HeaderExpectCT                        = "Expect-CT"
	HeaderFeaturePolicy                   = "Feature-Policy"
	HeaderPublicKeyPins                   = "Public-Key-Pins"
	HeaderPublicKeyPinsReportOnly         = "Public-Key-Pins-Report-Only"
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderUpgradeInsecureRequests         = "Upgrade-Insecure-Requests"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXDownloadOptions                = "X-Download-Options"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderXPoweredBy                      = "X-Powered-By"
	HeaderXXSSProtection                  = "X-XSS-Protection"

	// Server-sent event
	HeaderLastEventID = "Last-Event-ID"
	HeaderNEL         = "NEL"
	HeaderPingFrom    = "Ping-From"
	HeaderPingTo      = "Ping-To"
	HeaderReportTo    = "Report-To"

	// Transfer coding
	HeaderTE               = "TE"
	HeaderTrailer          = "Trailer"
	HeaderTransferEncoding = "Transfer-Encoding"

	// WebSockets
	HeaderSecWebSocketAccept     = "Sec-WebSocket-Accept"
	HeaderSecWebSocketExtensions = "Sec-WebSocket-Extensions" /* #nosec G101 */
	HeaderSecWebSocketKey        = "Sec-WebSocket-Key"
	HeaderSecWebSocketProtocol   = "Sec-WebSocket-Protocol"
	HeaderSecWebSocketVersion    = "Sec-WebSocket-Version"

	// Other
	HeaderAcceptPatch         = "Accept-Patch"
	HeaderAcceptPushPolicy    = "Accept-Push-Policy"
	HeaderAcceptSignature     = "Accept-Signature"
	HeaderAltSvc              = "Alt-Svc"
	HeaderDate                = "Date"
	HeaderIndex               = "Index"
	HeaderLargeAllocation     = "Large-Allocation"
	HeaderLink                = "Link"
	HeaderPushPolicy          = "Push-Policy"
	HeaderRetryAfter          = "Retry-After"
	HeaderServerTiming        = "Server-Timing"
	HeaderSignature           = "Signature"
	HeaderSignedHeaders       = "Signed-Headers"
	HeaderSourceMap           = "SourceMap"
	HeaderUpgrade             = "Upgrade"
	HeaderXDNSPrefetchControl = "X-DNS-Prefetch-Control"
	HeaderXPingback           = "X-Pingback"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderXRobotsTag          = "X-Robots-Tag"
	HeaderXUACompatible       = "X-UA-Compatible"
)
