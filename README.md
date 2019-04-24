## Guesser
Guesser is a Universal char by char guesser for Injections Attacks (and probably more).

Obviously, it does not pretend to be an alternative to other amazing tools that work really good exploiting these kind of attacks. However, from time to time I find scenarios where those tools doesn't work as expeted and I needed to code my own tools. Guesser was a way to have a common engine for all of them and be able to exploit these issues reusing as much code as possible.

## How to use Guesser

### Preparing guess script

First thing you need to do is to create a script that perform a guess. You can use curl, dig or whatever command is available in your platform.

An example of a guess script is following:

```bash
read OUT

echo "beef1234cafe
faa123" | grep "$OUT" 2>&1 >/dev/null

echo $?
```

This curl script has tree parts:
1. Read the term to be tested (`read` statement).
2. Execute the command that performs the guess (`curl`, `grep` or whatever), using the term provided above. We can use all our scripting ninjutsu at this point. We can use `curl`, `grep`, `time`, `wc` or whatever that let us know the difference between a right guess and a wrong guess. It's important that this command *should not print any result*.
3. Echo a *numeric value* (very important) depending on the result of the previous command.

### Running Guesser (examples)

Find terms in the charset [0-9]+

```$ guesser -charset "0123456789" -cmd "sh guess_script.sh"```

Find terms specifying a right and a wrong term samples (for comparison) different from the defaults

```$ guesser -right "_" -wrong "^" -cmd "sh guess.sh"```

Decrease the number of concurrent threads and add a delay between guesses

```$ guesser -threads 1 -delay 3000 -cmd "sh guess.sh"```

Run debug mode, when the tools doesn't work and we want to find why

```$ guesser -debug -cmd "sh guess.sh"```