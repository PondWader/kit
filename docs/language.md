# The Kit Language

The Kit language (kitlang) is used for repository/package configuration. It is programmable. The aim is to be as readable as possible with no prior experience of the language (but some programming experience). This means that unique features will not be used.

#### Naming conventions

All values & functions should be named with snake_case.

#### Types

The language is dynamically typed since it is designed for use in small scripts.

#### Top-level & function scope

In the top level scope all syntax is acceptable except function calls.

> **Why?** Because Kit is designed to be used as a configuration language and files should be able to be loaded without any side effects.

`export` can be placed before any declaration that should be exported and are only allowed in the top-level. Currently there is no way to import from another file so `export` only exists to make values available to the host process. For example, if you have a Kit file that describes a package it might contain `export name = "package-name"`.

#### Functions

Functions are only allowed to take a **single** argument and return a **single value**. The reasoning for this is multiple arguments can impaire readability of the code. It is better to accept an object, list or use method chaining. For example `min(a, b)` could be more clear as `min([a, b])` which shows the reader that the values are not doing unique things. `split("a.b.c", ".")` is more understandable when written as `split("a.b.c").at(".")`. This could be implemented using an implicit object return:

```
fn split(str) -> {
    fn at(sep) -> {
        // ...
        return parts
    }
}
```

By using the `->` whatever comes after will be returned.

#### Objects

All object properties are immutable. The values can be updated but keys cannot be added/removed. If you wish to dynamically add/remove key-value pairs you should use a map.

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

Strings are wrapped in regular quotation marks (`"like this"`). All strings can be used as template literals by wrapping variables in `${}`. For example ``

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

#### Memory management

All heap allocated types by default use an ownership model. However, if the value is being stored somewhere, 
such as in a list with a different lifetime the memory model must be declared.

For example:
```
my_list = [1, 2, 3]
another_list = [my_list] // This is not allowed!
another_list_2 = [RefCounted(my_list)]
```
