/*
The qs package can convert structs into query strings and vice versa.
The interface of qs is very similar to that of some standard marshaler
packages like "encoding/json", "encoding/xml".

Note that html forms are often POST-ed in the HTTP request body in the same format
as query strings (which is an encoding called application/x-www-form-urlencoded)
so this package can be used for that as well.
*/
package qs
