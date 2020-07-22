# Example::S3::Bucket AnalyticsConfiguration

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#tagfilters" title="TagFilters">TagFilters</a>" : <i>[ <a href="tagfilter.md">TagFilter</a>, ... ]</i>,
    "<a href="#storageclassanalysis" title="StorageClassAnalysis">StorageClassAnalysis</a>" : <i><a href="storageclassanalysis.md">StorageClassAnalysis</a></i>,
    "<a href="#id" title="Id">Id</a>" : <i>String</i>,
    "<a href="#prefix" title="Prefix">Prefix</a>" : <i>String</i>
}
</pre>

### YAML

<pre>
<a href="#tagfilters" title="TagFilters">TagFilters</a>: <i>
      - <a href="tagfilter.md">TagFilter</a></i>
<a href="#storageclassanalysis" title="StorageClassAnalysis">StorageClassAnalysis</a>: <i><a href="storageclassanalysis.md">StorageClassAnalysis</a></i>
<a href="#id" title="Id">Id</a>: <i>String</i>
<a href="#prefix" title="Prefix">Prefix</a>: <i>String</i>
</pre>

## Properties

#### TagFilters

_Required_: No

_Type_: List of <a href="tagfilter.md">TagFilter</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### StorageClassAnalysis

_Required_: Yes

_Type_: <a href="storageclassanalysis.md">StorageClassAnalysis</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Id

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Prefix

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
