Creating the Signature
----------------------
0) Every URL must supply:
  * public_key
  * timestamp with timezone (strictly RFC33339 format)
  * signature (computed below)

1) Create the canonicalized query string that you need later in this procedure:
  * Sort the query string components by parameter name with natural byte ordering.
  * URL encode the parameter name and values according to the following rules:
  * Do not URL encode any of the unreserved characters that RFC 3986 defines.
  * These unreserved characters are A-Z, a-z, 0-9, hyphen ( - ), underscore ( _ ), period ( . ), and tilde ( ~ ).
  * Percent encode all other characters with %XY, where X and Y are hex characters 0-9 and uppercase A-F.
  * Percent encode the space character as %20 (and not +, as common encoding schemes do).
  * Separate the encoded parameter names from their encoded values with the equals sign ( = )
  * Separate the name-value pairs with an ampersand ( & )

2) Create the string to sign according to the pseudo-grammar shown below (the "\n" represents an ASCII newline character).

```
StringToSign = HTTPRequestPath + "\n" + CanonicalizedQueryString <from the preceding step>
```

3) Calculate an RFC 2104-compliant HMAC with the string you just created, your Secret Access Key as the key, and SHA512 as the hash algorithm.

4) Convert the resulting value to base64

5) Use the resulting value as the value of the signature request parameter.
