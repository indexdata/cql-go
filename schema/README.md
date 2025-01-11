# Schema files

Here you'll find [`xcql.xsd`](xcql.xsd) which is identical to
https://docs.oasis-open.org/search-ws/searchRetrieve/v1.0/os/schemas/xcql.xsd
except for a change for sort key modifiers to be optional (like it is for
other modifiers).

It's also weird that the boolean wrapper element is `Boolean` (capital B),
unlike all other elements in the spec. In SRU 1.1 it was lower-case.
