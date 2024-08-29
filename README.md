# authplaceholder
Quick-and-dirty authentication placeholder I can drop into various hobby project until they get real auth.

## What is it?

The idea is to just get some quick access blocking thingmabobber in place, so the focus can be on prototyping the thing I'm building, rather than having to limit access to it. For example, if it stores a bunch of data from the user, I don't want some crawler to accidentally fill it with gibberish, but I also don't want to waste prototyping time locking that down.

Hence the placeholder authentication that basically just takes a username/password pair and sets a token as a cookie if you guess them right.

## Is it secure?

Absolutely not. The token is guessable. This is basically the equivalent of hanging a "No trespassers" sign in terms of security. Anyone that wants to can probably just wade on through it.

This is not security software. This is not proper authentication. Using this in production is dumb. Don't.

## How do I use it?

First, you make some JSON.

[The example JSON](example/example.json) has all the default settings, except the stylesheet and the users. The password encrypted there is `demon`.

You can use any bcrypt enabled software to generate passwords. I used [bcrypt-generator.com](https://bcrypt-generator.com/) for my testing.

It also has the very minimal code needed to get it spinning, in [example.go](example/example.go).

## Sweet, so when all my secrets get stolen, I can sue you for negligence?

No, but if you use this to "secure" someone else's secrets, they can sue *you* for negligence, so that's fun.

# Enjoy!

The intent here is to lower the bar for certain prototypes, so I hope that function is well served.