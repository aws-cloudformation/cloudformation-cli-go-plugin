# Example::S3::Bucket Rule

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#status" title="Status">Status</a>" : <i>String</i>,
    "<a href="#noncurrentversionexpirationindays" title="NoncurrentVersionExpirationInDays">NoncurrentVersionExpirationInDays</a>" : <i>Double</i>,
    "<a href="#transitions" title="Transitions">Transitions</a>" : <i>[ <a href="transition.md">Transition</a>, ... ]</i>,
    "<a href="#tagfilters" title="TagFilters">TagFilters</a>" : <i>[ <a href="tagfilter.md">TagFilter</a>, ... ]</i>,
    "<a href="#noncurrentversiontransitions" title="NoncurrentVersionTransitions">NoncurrentVersionTransitions</a>" : <i>[ <a href="noncurrentversiontransition.md">NoncurrentVersionTransition</a>, ... ]</i>,
    "<a href="#prefix" title="Prefix">Prefix</a>" : <i>String</i>,
    "<a href="#noncurrentversiontransition" title="NoncurrentVersionTransition">NoncurrentVersionTransition</a>" : <i><a href="noncurrentversiontransition.md">NoncurrentVersionTransition</a></i>,
    "<a href="#expirationdate" title="ExpirationDate">ExpirationDate</a>" : <i>String</i>,
    "<a href="#expirationindays" title="ExpirationInDays">ExpirationInDays</a>" : <i>Double</i>,
    "<a href="#transition" title="Transition">Transition</a>" : <i><a href="transition.md">Transition</a></i>,
    "<a href="#id" title="Id">Id</a>" : <i>String</i>,
    "<a href="#abortincompletemultipartupload" title="AbortIncompleteMultipartUpload">AbortIncompleteMultipartUpload</a>" : <i><a href="abortincompletemultipartupload.md">AbortIncompleteMultipartUpload</a></i>
}
</pre>

### YAML

<pre>
<a href="#status" title="Status">Status</a>: <i>String</i>
<a href="#noncurrentversionexpirationindays" title="NoncurrentVersionExpirationInDays">NoncurrentVersionExpirationInDays</a>: <i>Double</i>
<a href="#transitions" title="Transitions">Transitions</a>: <i>
      - <a href="transition.md">Transition</a></i>
<a href="#tagfilters" title="TagFilters">TagFilters</a>: <i>
      - <a href="tagfilter.md">TagFilter</a></i>
<a href="#noncurrentversiontransitions" title="NoncurrentVersionTransitions">NoncurrentVersionTransitions</a>: <i>
      - <a href="noncurrentversiontransition.md">NoncurrentVersionTransition</a></i>
<a href="#prefix" title="Prefix">Prefix</a>: <i>String</i>
<a href="#noncurrentversiontransition" title="NoncurrentVersionTransition">NoncurrentVersionTransition</a>: <i><a href="noncurrentversiontransition.md">NoncurrentVersionTransition</a></i>
<a href="#expirationdate" title="ExpirationDate">ExpirationDate</a>: <i>String</i>
<a href="#expirationindays" title="ExpirationInDays">ExpirationInDays</a>: <i>Double</i>
<a href="#transition" title="Transition">Transition</a>: <i><a href="transition.md">Transition</a></i>
<a href="#id" title="Id">Id</a>: <i>String</i>
<a href="#abortincompletemultipartupload" title="AbortIncompleteMultipartUpload">AbortIncompleteMultipartUpload</a>: <i><a href="abortincompletemultipartupload.md">AbortIncompleteMultipartUpload</a></i>
</pre>

## Properties

#### Status

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### NoncurrentVersionExpirationInDays

_Required_: No

_Type_: Double

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Transitions

_Required_: No

_Type_: List of <a href="transition.md">Transition</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### TagFilters

_Required_: No

_Type_: List of <a href="tagfilter.md">TagFilter</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### NoncurrentVersionTransitions

_Required_: No

_Type_: List of <a href="noncurrentversiontransition.md">NoncurrentVersionTransition</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Prefix

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### NoncurrentVersionTransition

_Required_: No

_Type_: <a href="noncurrentversiontransition.md">NoncurrentVersionTransition</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ExpirationDate

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ExpirationInDays

_Required_: No

_Type_: Double

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Transition

_Required_: No

_Type_: <a href="transition.md">Transition</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Id

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### AbortIncompleteMultipartUpload

_Required_: No

_Type_: <a href="abortincompletemultipartupload.md">AbortIncompleteMultipartUpload</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
