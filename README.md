# Custom Pod Autoscaler

This is the Custom Pod Autoscaler (CPA) code and base images.

## Use

The CPA can be used to make Customised Autoscalers, either by using one of the base images or building a new base image with the binary. See the example folder for examples.

## Developing
### Environment
Developing this project requires these dependencies:

* Go >= 1.13

### Commands

* make - builds the CPA binary.
* make docker - builds the CPA base images.
* make lint - lints the code.
* make test - tests the code.
* make vendor - generates a vendor folder.