# Example::S3::Bucket Transition

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#transitiondate" title="TransitionDate">TransitionDate</a>" : <i>String</i>,
    "<a href="#transitionindays" title="TransitionInDays">TransitionInDays</a>" : <i>Double</i>,
    "<a href="#storageclass" title="StorageClass">StorageClass</a>" : <i>String</i>
}
</pre>

### YAML

<pre>
<a href="#transitiondate" title="TransitionDate">TransitionDate</a>: <i>String</i>
<a href="#transitionindays" title="TransitionInDays">TransitionInDays</a>: <i>Double</i>
<a href="#storageclass" title="StorageClass">StorageClass</a>: <i>String</i>
</pre>

## Properties

#### TransitionDate

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### TransitionInDays

_Required_: No

_Type_: Double

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### StorageClass

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

