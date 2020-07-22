# Example::S3::Bucket InventoryConfiguration

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#destination" title="Destination">Destination</a>" : <i><a href="destination.md">Destination</a></i>,
    "<a href="#optionalfields" title="OptionalFields">OptionalFields</a>" : <i>[ String, ... ]</i>,
    "<a href="#includedobjectversions" title="IncludedObjectVersions">IncludedObjectVersions</a>" : <i>String</i>,
    "<a href="#enabled" title="Enabled">Enabled</a>" : <i>Boolean</i>,
    "<a href="#id" title="Id">Id</a>" : <i>String</i>,
    "<a href="#prefix" title="Prefix">Prefix</a>" : <i>String</i>,
    "<a href="#schedulefrequency" title="ScheduleFrequency">ScheduleFrequency</a>" : <i>String</i>
}
</pre>

### YAML

<pre>
<a href="#destination" title="Destination">Destination</a>: <i><a href="destination.md">Destination</a></i>
<a href="#optionalfields" title="OptionalFields">OptionalFields</a>: <i>
      - String</i>
<a href="#includedobjectversions" title="IncludedObjectVersions">IncludedObjectVersions</a>: <i>String</i>
<a href="#enabled" title="Enabled">Enabled</a>: <i>Boolean</i>
<a href="#id" title="Id">Id</a>: <i>String</i>
<a href="#prefix" title="Prefix">Prefix</a>: <i>String</i>
<a href="#schedulefrequency" title="ScheduleFrequency">ScheduleFrequency</a>: <i>String</i>
</pre>

## Properties

#### Destination

_Required_: Yes

_Type_: <a href="destination.md">Destination</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### OptionalFields

_Required_: No

_Type_: List of String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### IncludedObjectVersions

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Enabled

_Required_: Yes

_Type_: Boolean

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Id

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Prefix

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ScheduleFrequency

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

