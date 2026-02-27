# The Kit Language

The Kit language (kitlang) is used for repository/package configuration. It is
programmable. The aim is to be as readable as possible with no prior experience
of the language (but some programming experience). This means that unique
features will not be used.

#### Naming conventions

All values & functions should be named with snake_case.

#### Types

The language is dynamically typed since it is designed for use in small scripts.

#### Top-level & function scope

In the top level scope all syntax is acceptable except function calls.

> **Why?** Because Kit is designed to be used as a configuration language and
> files should be able to be loaded without any side effects.

`export` can be placed before any declaration that should be exported and are
only allowed in the top-level. Currently there is no way to import from another
file so `export` only exists to make values available to the host process. For
example, if you have a Kit file that describes a package it might contain
`export name = "package-name"`.

#### Functions

Functions are only allowed to take a **single** argument and return a **single
value**. The reasoning for this is multiple arguments can impaire readability of
the code. It is better to accept an object, list or use method chaining. For
example `min(a, b)` could be more clear as `min([a, b])` which shows the reader
that the values are not doing unique things. `split("a.b.c", ".")` is more
understandable when written as `split("a.b.c").at(".")`. This could be
implemented using an implicit object return:

```
fn split(str) -> {
    fn at(sep) -> {
        // ...
        return parts
    }
}
```

By using the `->` whatever comes after will be returned.

Functions can also take a type check such as `fn split(str: string)`. This is
syntactic sugar for performing **runtime** checking of the type and returning an
error if it doesn't match.

Supported function arg type annotations are currently:

```kit
string
bool
number
```

Functions can also destructure a single object argument:

```kit
fn load_component_packages({ url, suite, component, arch }) {
    index_url = "${url}/dists/${suite}/${component}/binary-${arch}/Packages.xz"
    // ...
}
```

Destructuring currently supports key names only (no defaults or renaming).

#### Objects

All object properties are immutable. The values can be updated but keys cannot
be added/removed. If you wish to dynamically add/remove key-value pairs you
should use a map.

Example declaration of an object:

```kit
pet = {
    name = "Milo"
    type = "dog"
    birthday = date("2023-07-13")
    age = calc_age("")
}
```

#### Lists

Lists store a dynamic collection of values of any type.

A list literal can be declared like so:

```kit
pets = [pet1, pet2]
first_per = pets[0] // Lists are 0 indexed
```

Bracket indexing (`value[index]`) works on indexable values and uses runtime
index checks.

Lists can be mutated:

```
pets.append(pet)
pets.prepend(pet)
pets.remove_first()
pets.remove_last()
pets.remove_at(3)

pets.length()
```

Additional methodlly lists contain all the methods from streams.

#### Streams

Streams is a general term that represents all types that can be iterated over.

Stream methods:

```
map()
map_to_set()
filter()
```

#### Strings

Strings are wrapped in regular quotation marks (`"like this"`). All strings can
be used as template literals by wrapping variables in `${}`. For example ``

#### Classes

There are no classes as such. But have a look at this syntax

```
fn Person(person) -> {
    name = person.name
    date_of_birth = person.data_of_birth

    fn age() {
        // ... Calculate age
        return years
    }
}
```

#### Errors and throw

Errors are nominal interface instances. The standard library provides:

```kit
Error          // Interface reference
error(message) // Constructor function, returns an Error instance
```

Use `throw` to return an error:

```kit
throw error("could not find package")
```

`throw` requires its argument to be an instance of the standard `Error`
interface.

#### Interfaces

Interfaces can contain fields and methods and can be declared like so:

```kit
interface Stream {
    bytes_read: number
    fn read() 
}
```

An object can be declared to implement an interface by using the `instance`
keyword before the interface reference:

```kit
stream = instance Stream {
    bytes_read = 0

    fn read() {
        return [1, 2, 3]
    }

    // Extra fields/methods are allowed
    source = "stdin"
    fn close() {
        return nil
    }
}
```

An object must be created via `instance InterfaceName { ... }` to be acceptable
as a type of that interface.

#### Memory management (draft)

All heap allocated types by default use an ownership model. However, if the
value is being stored somewhere, such as in a list with a different lifetime the
memory model must be declared.

For example:

```kit
my_list = [1, 2, 3]
another_list = [my_list] // This is not allowed!
another_list_2 = [RefCounted(my_list)]
```

#### Async/threading (draft)

Future is a core type that means a function will return a value in the future.
Future always immediately returns whilst other execution can continue

Functions can be marked with `async` or `fork`. `fork` runs the code in
potentially another thread (green threads managed by the Kit runtime) whilst
`async` just means other actions can be performed whilst i/o is pending.

`fork` should always be preferred as it allows the program to leverage multiple
threads however `async` can be used for thread-safety reasons when writing
programs that operate on the same memory.

```
fork fn connect(): Future<Conn> {
    conn = (fork tcp.connect("0.0.0.0",443)).await()
    conn = connect().await()
   return tcp.connect("0.0.0.0", 443) 
} 

// A warning should be emitted that this function could use fork since it has no side effect
async fn connect(): Future<Conn> {
   return async tcp.connect("0.0.0.0", 443) 
}
```
