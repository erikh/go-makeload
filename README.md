## makeload: a small library to make HTTP load

I haven't written code in some time and needed to warm up. So I wrote this in preparation for a project that needs it, and eventually as I improve this library, this message will disappear and more information about what it actually does will appear in its place.

This library is a HTTP load generator, similar in function to `wrk` or `ab`. It has a programmable interface, and is intended for use in integration testing of a HTTP service. It has a very small statistics collector, an interface for delivering requests, and concurrency and connection controls. It is currently very focused on lean functionality and accuracy. The library has reliably tested to deliver the exact amount of requests you deliver it, saturating the exact number of cores fitting the concurrency mark.

Please see the tests for examples of how to use this library. I also feel the code and interfaces are simple enough that most experienced Golang programmers should need little instruction using it.


### The future

As mentioned, more is to come with regards to this library's functionality. Here are some things that will probably show up eventually:

- [ ] Statistics for mean delivery time
- [ ] Programmable functionality for determining errors / valid responses (right now just non-200's are errors)
- [ ] Programmable request delivery
- [ ] Documentation (thankfully, right now it's very small)
- [ ] Some more self-testing

As mentioned, I'm shipping this to be a part of another product. If you file bugs for it, I will attempt to service requests, but if they conflict with the other project's goals, I strongly suggest you fork this library instead of push harder for your changes, which is MIT licensed for a reason.

May peace be with you.

### Author

Erik Hollensbe <erik+github@hollensbe.org>
