# Parallel file copy and checksum

### Note

The `Go` code is basically the same as what I wrote for work but this README is
sanitized and expanded for public consumption and education. This includes an
explanation of parallel vs concurrent for completeness.

## Use case

I want to copy a large number of files (say 60+) with sizes varying from 100 MB
to 30+ GB; totalling 800+ GB. If I copy them using the `cp` or `tar` commands,
it takes a rather long time because I'm copying them one at a time (serially).
If I can batch them into 10 groups and copy the 10 groups at the same time
things could be faster.

I tried this tactic using a shell script that used `cp` to copy each file listed
in a file and ran the script with 10 different files in the background.
Something like:

```
$ for i in $(01 02 03 04 05 06 07 08 09 10); do nohup ./mcopy.sh f${i} >f${i}.out 2>&1 &; done
```

This was faster but there was manual work to create the files with lists of
files in them. Of course, I could have scripted that, too. But that's too many
scripts and things for others to wrap their heads around; let alone keep
straight.

Instead, to practice writing programs in `Go` and to get familiar with
`goroutines`, I used `Go` to implement this tactic of copying many files at the
same time from one directory to another. Better for others, there is one program
to run and they can just point it at a source and destination directory. _Easy
peasy!_

Also, I wanted to make sure that the bits didn't get mangled in the copy
process, so I needed to compare the checksum of the original files against the
checksums of the copies. So, I wrote another program to compute the checksums of
files in a directory and do that at the same time.

## The programs:

* `pcopy` - copy files from a source directory to a destination directory and
  compute the SHA1 checksum of each file while copying. Think `cat file1 | tee
  $DEST_DIR/file1 | sha1sum`.
* `psha1sum` - compute the SHA1 checksum of each file in a directory

Both programs use `Go`'s concurrent programming support to do their work with up
to 30 concurrent workers. The default is 10 workers. These programs speed up the
tasks of copying files from one directory to another and computing the SHA1
checksum of files in a directory. Only the files in the top level of a directory
are affected. In other words, files in subdirectories are not copied or
checksummed.

## Example use

How do I run the programs?

```
$ pcopy
usage: pcopy [ -w ] srcDir destDir

$ psha1sum
usage: psha1sum [ -w ] dir
```

Copy files from a source directory to a destination directory with 30 workers.

```
$ pcopy -w 30 /mydb/prod /db-backup/2018-12-22
```

Compute the checksum of files in a directory with 5 workers

```
$ psha1sum -w 5 /db-backup/2018-12-22
```

Then, compare the checksums from the `pcopy` output to those from the `psha1sum`
output to verify that the copies are identical to the originals.

## CAVEATS

The amount of speed up depends on the number of CPUs of the server and the
throughput of the disk subsystem. In the typical, virtual server environment
with disk arrays providing shared storage, the speed up can vary widely from day
to day and even hour to hour. There's just no guarantee in a virtualized (i.e.,
shared) environment.

## Building the programs

The convention with `Go` program development is your _**workspace**_ is the top
level directory where you have a `src`, `pkg`, and `bin` directory and in the `src`
directory you have subdirectories for your programs. See [How to Write Go
Code](https://golang.org/doc/code.html).

Your `GOPATH` environment variable is set to your _**workspace**_ directory.

### Clone the repo

```
$ cd $GOPATH/src

$ git clone $GITHUB_REPO_URL

```

where `GITHUB_REPO_URL` is the URL from the `Clone or download` button.

### On Linux:

```
$ cd $GOPATH/src/parallel-file-copy-and-checksum

$ go build pcopy.go
$ go build psha1sum.go
```

### On macos:

We can cross-compile for a Linux target.

```
$ cd $GOPATH/src/parallel-file-copy-and-checksum/

$ GOOS=linux go build pcopy.go
$ GOOS=linux go build psha1sum.go
```

### On Windows:

It should work but I did not try it to verify that it works.


## Installing the programs

Copy the executables, `pcopy` and `psha1sum`, to /usr/local/bin on the target
Linux server. Make sure that they are Linux executables (ELF-64).

## Parallel vs Concurrent

In his [Concurrency is not
Parallelism](https://www.youtube.com/watch?v=oV9rvDllKEg) talk, Rob Pike says,
> _"**concurrency** is the composition of independently executing processes"_
> while _"**parallelism**, on the other hand, is the simultaneous execution of
> multiple things; possibly related, possibly not"_

The primary difference between parallel execution and concurrent execution is
parallel means at the same time while concurrent means interleaved over a period
of time. It's convenient for the average person to think the two are the same
but they are not. The subtle difference may not make a big difference in most
cases. So, from a larger context treating them the same is probably OK. Just be
aware.

See:

* [Can you have concurrency but not
  parallelism?](https://github.com/Criviere/os/blob/master/Chapter4.md)
* [Parallel Programming vs. Concurrent
  Programming](https://takuti.me/note/parallel-vs-concurrent/)
* [Concurrency is not Parallelism](https://www.youtube.com/watch?v=oV9rvDllKEg)
  - by Rob Pike from Heroku's Waza conference in January 2012,
    [slides](https://talks.golang.org/2012/waza.slide#1)
