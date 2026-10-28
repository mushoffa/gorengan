# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.13.0](https://github.com/mushoffa/gorengan/compare/v0.12.2...v0.13.0) (2026-10-28)



### Features

* **fyne:** Add bar and line chart padding options and app theme support ([380cee0](https://github.com/mushoffa/fatamorgana/commit/380cee0c25deb83e381e6c0f38182ccd82d06fd6))
* **fyne:** Add chart padding options, show axis options, and app theme support ([b54fdc5](https://github.com/mushoffa/fatamorgana/commit/b54fdc5e3c7eaf30da14f1c51f9ad9b318bdbec4))
* **widget:** Add Surface widget ([0b9851f](https://github.com/mushoffa/fatamorgana/commit/0b9851fa5d314d374ee864ec236e78260d17462f))
* **fyne:** Add ShowXAxis and ShowYAxis to hide line and bar chart axis ([4bb1f2e](https://github.com/mushoffa/fatamorgana/commit/4bb1f2e7089c396b51b5289c5dd69a4369d55042))
* **widget:** Add ImageViewer with zoom, compare and built-in overlay ([75c8af2](https://github.com/mushoffa/fatamorgana/commit/75c8af2bb43bfef125113906c5afada76baa7518))
* **fyne:** Add OpenImageButton with built-in file dialog ([6ab94c1](https://github.com/mushoffa/fatamorgana/commit/6ab94c1436c415c5814ac9dec03cbe3a9ba95aa3))
* **fyne:** Add SidePanel with optional label, header, scrollable body and footer ([14f9e61](https://github.com/mushoffa/fatamorgana/commit/14f9e61117c358cf129e0e4b482688978ff96061))

## [0.12.2](https://github.com/mushoffa/gorengan/compare/v0.12.1...v0.12.2) (2026-10-24)



### Bug Fixes

* **fyne:** Make Liena and Bar chart SetData fit axis ranges and accept nil x ([8a565fe](https://github.com/mushoffa/fatamorgana/commit/8a565feebae13cf7ed8ee8eca1fd93f17e80234e))

## [0.12.1](https://github.com/mushoffa/gorengan/compare/v0.12.0...v0.12.1) (2026-10-24)



### Bug Fixes

* **widget:** Replace smallTextSize() with a theme.TextSize() constant in CompactCheck ([903cc15](https://github.com/mushoffa/fatamorgana/commit/903cc154634450450dd293c741537e33a47c4f8f))

## [0.12.0](https://github.com/mushoffa/gorengan/compare/v0.11.0...v0.12.0) (2026-10-24)



### Features

* **widget:** Add default chart theme ([b2af3a5](https://github.com/mushoffa/fatamorgana/commit/b2af3a534457b4d58652b9ce7e33ab1826fcea6c))

## [0.11.0](https://github.com/mushoffa/gorengan/compare/v0.10.0...v0.11.0) (2026-10-24)



### Features

* **chart:** Add interactive bar and line charts with compact Fill check ([60c808e](https://github.com/mushoffa/fatamorgana/commit/60c808e8b4b761245b9a2217c4263d03057952f7))

## [0.10.0](https://github.com/mushoffa/gorengan/compare/v0.9.0...v0.10.0) (2026-10-24)



### Features

* **http:** Add unified WebClient with net/http, Fiber V2, and Fiber V3 ([e7a1990](https://github.com/mushoffa/fatamorgana/commit/e7a19907e6154692847d4c80855c6794202037d2))
* **fyne:** Add NumericEntry with label, data binding and optional theme ([9e4b6f2](https://github.com/mushoffa/fatamorgana/commit/9e4b6f264ec79ed808514938d31f5d048e3b7596))


### Tests

* **http:** Add unit tests for the client package ([5f1a898](https://github.com/mushoffa/fatamorgana/commit/5f1a89836d6dac792cd018a1ba10b35414975847))

## [0.9.0](https://github.com/mushoffa/gorengan/compare/v0.8.0...v0.9.0) (2024-10-12)


### Image

* **Monochrome:** Add inverse function and param ([d359354](https://github.com/mushoffa/gorengan/commit/d35935441d21c7a91b78af7ae5290236d2a2bb7c))


### Tests

* **image:** Update monochrome input param on unit test ([fc928ce](https://github.com/mushoffa/gorengan/commit/fc928ce3a46cff98d6f328e9695a30e6817c045c))

## [0.8.0](https://github.com/mushoffa/gorengan/compare/v0.7.2...v0.8.0) (2024-08-06)


### Minio

* **ListFolder:** Fix previous commit on spliting prefix name ([6cafc91](https://github.com/mushoffa/gorengan/commit/6cafc914051add5ce90e87baf7435965077ae3bd))
* **ListFolders:** Change response type to JSON string ([c82974f](https://github.com/mushoffa/gorengan/commit/c82974fd174e6ddb4302d16564cb241331477ff0))

### [0.7.2](https://github.com/mushoffa/gorengan/compare/v0.7.1...v0.7.2) (2024-08-01)


### Bug Fixes

* **image:** Expose field Data of Image struct ([edacc4e](https://github.com/mushoffa/gorengan/commit/edacc4e7373c38929734bf3e336d9694b8f797ba))

### [0.7.1](https://github.com/mushoffa/gorengan/compare/v0.7.0...v0.7.1) (2024-08-01)


### Bug Fixes

* **Image:** Push forgotten file(s) ([08ee3e3](https://github.com/mushoffa/gorengan/commit/08ee3e32304c7194f517c69d92b96b95bb004e19))

## [0.7.0](https://github.com/mushoffa/gorengan/compare/v0.6.0...v0.7.0) (2024-08-01)


### Documentation

* **readme:** Add goreport badge ([1b3f8ff](https://github.com/mushoffa/gorengan/commit/1b3f8ff8fceaf2ef1ee816c210457cf4a0d00f3c))


### Image

* Add crop function for Image and Monochrome type ([7dc2fab](https://github.com/mushoffa/gorengan/commit/7dc2fabaeb35df7e01249fe5e7132da8d0ebccf1))

## [0.6.0](https://github.com/mushoffa/gorengan/compare/v0.5.0...v0.6.0) (2024-07-29)


### Liquibase

* Add migration template and deployment strategy ([abe34a8](https://github.com/mushoffa/gorengan/commit/abe34a8758b05f270306f3d0d6612a1ac09f181e))


### Documentation

* Add Liquibase std version section ([908af69](https://github.com/mushoffa/gorengan/commit/908af69f3c5afd813c711c9a6585f1654c86df00))
* **minio:** Change readme filename ([59b7bf1](https://github.com/mushoffa/gorengan/commit/59b7bf165bd3fea3055d52e2155b47b5ae0e0cd7))
* **readme:** Add badge ([c64e21a](https://github.com/mushoffa/gorengan/commit/c64e21a19f7e7219c31c568027ca53818ea06f1d))

## [0.5.0](https://github.com/mushoffa/gorengan/compare/v0.4.0...v0.5.0) (2024-07-22)


### Image

* **monochrome:** Enhance monochrome process with concurrent process ([afe8b1e](https://github.com/mushoffa/gorengan/commit/afe8b1e094765b500e80c561c700bb9b0ff78223))


### Build System

* Add minio on standard version section ([e8fd44a](https://github.com/mushoffa/gorengan/commit/e8fd44a26eb3934f9f1a37dcdd387e3b4e81c9ce))
* Bump dependency ([b025a5c](https://github.com/mushoffa/gorengan/commit/b025a5c26f210d349855f21b74743ab6a0d8e154))


### Minio

* New library ([e2fbbaa](https://github.com/mushoffa/gorengan/commit/e2fbbaae962294315a3646989a0228e23b4c50f0))


### Tests

* **image:** Add unit test, benchmark, and input test image ([d227efe](https://github.com/mushoffa/gorengan/commit/d227efeb3e38b1dfd671cc96a2344fd62a980b0f))
* **minio:** Add makefile to run unit tests ([616952f](https://github.com/mushoffa/gorengan/commit/616952f33355293cd14b4bb4097e54eba96d58a9))
* **minio:** Add unit tests and sample input files ([bd4d34d](https://github.com/mushoffa/gorengan/commit/bd4d34d63663a576a319ae50fac020ca3e52e0c5))


### Documentation

* **minio:** Add readme ([225fce6](https://github.com/mushoffa/gorengan/commit/225fce64f9cb90f0a3ec0687f48e91332db1fc3b))

## [0.4.0](https://github.com/mushoffa/gorengan/compare/v0.3.1...v0.4.0) (2024-07-13)


### Image

* **monochrome:** Add function to convert to monochrome image ([e355179](https://github.com/mushoffa/gorengan/commit/e3551796b98ceed3edf0daf365c58eb1769f84e0))

### [0.3.1](https://github.com/mushoffa/gorengan/compare/v0.3.0...v0.3.1) (2024-07-13)


### Bug Fixes

* **histogram:** Fix vertical axis orientation ([82c73ce](https://github.com/mushoffa/gorengan/commit/82c73ce9831f5ab72e871fb0e0ed10633e8baf2d))

## [0.3.0](https://github.com/mushoffa/gorengan/compare/v0.2.0...v0.3.0) (2024-07-13)


### Image

* **histogram:** New function to generate histogram of monochrome image ([1d54a64](https://github.com/mushoffa/gorengan/commit/1d54a64b88f5e85e8db79414942cf45462bf3834))

## [0.2.0](https://github.com/mushoffa/gorengan/compare/v0.1.0...v0.2.0) (2024-07-13)


### Image

* **monochrome:** New library to convert color image to monochrome image ([3d806d8](https://github.com/mushoffa/gorengan/commit/3d806d828be85107226e9bff650ce30af43eaccc))
* **reader:** New reader to read image from file ([c4d3ba1](https://github.com/mushoffa/gorengan/commit/c4d3ba1330df352f92f85e0fcbd9e55216a12e89))
* **writer:** New writer to write image to file ([05358e8](https://github.com/mushoffa/gorengan/commit/05358e8c2b6fb76eb58eec21b315d5e46823e6d9))


### Tests

* **image:** Add image for testing purpose ([23d3e9b](https://github.com/mushoffa/gorengan/commit/23d3e9b2d2910ac4e480d82a092a1cdaa003b3c8))
* **monochrome:** Add unit test ([edd49c5](https://github.com/mushoffa/gorengan/commit/edd49c5c5102277d633947c17407b57d6ad1ff0b))


### Build System

* Add testify dependency library ([ccb9c66](https://github.com/mushoffa/gorengan/commit/ccb9c6606e5157c655a687ca6363a1e348560699))


### Documentation

* **readme:** Add Go reference badge ([f8f3466](https://github.com/mushoffa/gorengan/commit/f8f346615a8ab651d557dc15e7070e91c083fdf0))

## 0.1.0 (2024-07-12)


### Generic

* **slice:** Add new generic slice data type ([eb32e5a](https://github.com/mushoffa/gorengan/commit/eb32e5a17b224ce4e97c15e2ca3e4ff7a0e2df8f))


### Image

* Add new extension of std image pkg with additional conversion function ([aa0d60b](https://github.com/mushoffa/gorengan/commit/aa0d60b4a779c87bcd39fca9726d5195f6e475a7))


### Documentation

* **readme:** Add description and getting started stuffs ([e6cb9ed](https://github.com/mushoffa/gorengan/commit/e6cb9ed13c0ed26ff48a338b5516b187b3f1657d))


### Build System

* Add go.mod ([0fd1ed9](https://github.com/mushoffa/gorengan/commit/0fd1ed9c9c3cefdf7a10c672d0e3dd5d81562e92))
* Customize standard version section type ([601378d](https://github.com/mushoffa/gorengan/commit/601378dc709d89c2aef7a3e3ef7708243dd25c85))
