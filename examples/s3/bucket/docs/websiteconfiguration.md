# Example::S3::Bucket WebsiteConfiguration

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#routingrules" title="RoutingRules">RoutingRules</a>" : <i>[ <a href="routingrule.md">RoutingRule</a>, ... ]</i>,
    "<a href="#indexdocument" title="IndexDocument">IndexDocument</a>" : <i>String</i>,
    "<a href="#redirectallrequeststo" title="RedirectAllRequestsTo">RedirectAllRequestsTo</a>" : <i><a href="redirectallrequeststo.md">RedirectAllRequestsTo</a></i>,
    "<a href="#errordocument" title="ErrorDocument">ErrorDocument</a>" : <i>String</i>
}
</pre>

### YAML

<pre>
<a href="#routingrules" title="RoutingRules">RoutingRules</a>: <i>
      - <a href="routingrule.md">RoutingRule</a></i>
<a href="#indexdocument" title="IndexDocument">IndexDocument</a>: <i>String</i>
<a href="#redirectallrequeststo" title="RedirectAllRequestsTo">RedirectAllRequestsTo</a>: <i><a href="redirectallrequeststo.md">RedirectAllRequestsTo</a></i>
<a href="#errordocument" title="ErrorDocument">ErrorDocument</a>: <i>String</i>
</pre>

## Properties

#### RoutingRules

_Required_: No

_Type_: List of <a href="routingrule.md">RoutingRule</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### IndexDocument

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### RedirectAllRequestsTo

_Required_: No

_Type_: <a href="redirectallrequeststo.md">RedirectAllRequestsTo</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ErrorDocument

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
