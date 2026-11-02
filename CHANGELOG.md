# 1.0.0
* Require Go 1.23 or later and update dependencies.
* Preserve zero-allocation logging while escaping JSON and Logrus strings correctly.
* Fix rolling file period indexes, rotation deadlines, retries and archive collisions.
* Preserve source logs on compression failure and retain archive permissions.
* Fix archive retention progress and file handling during rotation failures.
* Fix asynchronous writer shutdown, restart and queue overflow behavior. Full queues discard all log levels without waiting for space.
* Synchronize logger levels and writer bindings, and harden socket, syslog, reader and provider lifecycles.
* Fix empty Pattern fields, literal commas, weekday formatting and timezone handling.
* Fix nil error handling, Zap unsigned integer fields and bridge panic severity mapping.
* Use a five-second default socket reconnection delay. ReconnectionDelay accepts time.Duration; specify units explicitly.
* Document encoder ownership: use an independent JSON or Pattern encoder for each writer.

# 0.7.0
* Add configurator to add manual config with new api.
* Add logger provider.
* Add classic provider to log more efficiently.
* Fix bug for writer in goroutine.
* Add bytes writer and event writer.

# 0.6.2
* Fix bug for filter when checking result.

# 0.6.1
* Add syslog writer implementation.
* Fix bug for async writer.

# 0.6.0
* Add builtin lork logger.
* Split project into multiple modules.
* Change write raw bytes to log event.
* Change `Interface` to `Any`.
* Change the return value of `Filter``.
* Fix bug.

# 0.5.4
* Fix alloc for atoi and `Msg`.
* Remove hex log.
* Add archiver and fix time format.

# 0.5.3
* Fix new line bug.
* Update abbreviation strategy for logger name.

# 0.5.2
* Fix LoggerC.

# 0.5.1
* Fix bug.

# 0.5.0
* Add log remover.

# 0.4.0
* Add sub logger.
* Add log event instead of byte array.
* Fix file writer.

# 0.3.0
* Add asynchronous writer and update socket writer.
* Reduce all allocation to zero. 

# 0.2.0
* Optimize pattern encoder and reduce the allocation.

# 0.1.0
* First implementation for lork.
