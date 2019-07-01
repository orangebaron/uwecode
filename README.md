# λuwecode
a functional language based entirely on pure [lambda calculus](https://en.wikipedia.org/wiki/Lambda_calculus)
## outline:
- no special types; numbers, booleans, strings, lists, etc. are based in pure lambda calculus
- IO system is based in pure lambda calculus
- type system is based in pure lambda calculus
- no keywords except for import directives
- inline functions are interchangeable with prefixed calling: ``a `b c`` == `b a c`
## types:
```
number: (x -> x) -> x -> x
bool: x -> x -> x
maybe a: (a -> x) -> x -> x
either a b: (a -> x) -> (b -> x) -> x
a,b: (a -> b -> x) -> x
[a]: maybe (a,[a])
byte: (((bool,bool),(bool,bool)),((bool,bool),(bool,bool)))
str: [byte]
IO: either (either (str -> IO) (str,IO)) (Maybe (IO,IO))
type: either number (type,type)
```
will describe more later
