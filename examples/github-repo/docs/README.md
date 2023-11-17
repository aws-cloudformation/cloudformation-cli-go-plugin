# Example::GitHub::Repo

Manages a GitHub Repo

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "Type" : "Example::GitHub::Repo",
    "Properties" : {
        "<a href="#name" title="Name">Name</a>" : <i>String</i>,
        "<a href="#owner" title="Owner">Owner</a>" : <i>String</i>,
        "<a href="#description" title="Description">Description</a>" : <i>String</i>,
        "<a href="#homepage" title="Homepage">Homepage</a>" : <i>String</i>,
        "<a href="#oauthtoken" title="OauthToken">OauthToken</a>" : <i>String</i>,
    }
}
</pre>

### YAML

<pre>
Type: Example::GitHub::Repo
Properties:
    <a href="#name" title="Name">Name</a>: <i>String</i>
    <a href="#owner" title="Owner">Owner</a>: <i>String</i>
    <a href="#description" title="Description">Description</a>: <i>String</i>
    <a href="#homepage" title="Homepage">Homepage</a>: <i>String</i>
    <a href="#oauthtoken" title="OauthToken">OauthToken</a>: <i>String</i>
</pre>

## Properties

#### Name

Name of the repository on GitHub

_Required_: Yes

_Type_: String

_Minimum Length_: <code>1</code>

_Maximum Length_: <code>50</code>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Owner

Where to create the repository, either a user or an organization

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Description

The title of the TPS report is a mandatory element.

_Required_: No

_Type_: String

_Minimum Length_: <code>20</code>

_Maximum Length_: <code>250</code>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Homepage

Homepage of the project

_Required_: No

_Type_: String

_Minimum Length_: <code>20</code>

_Maximum Length_: <code>250</code>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### OauthToken

OAuth token from GitHub

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

## Return Values

### Ref

When you pass the logical ID of this resource to the intrinsic `Ref` function, Ref returns the URL.

### Fn::GetAtt

The `Fn::GetAtt` intrinsic function returns a value for a specified attribute of this type. The following are the available attributes and sample return values.

For more information about using the `Fn::GetAtt` intrinsic function, see [Fn::GetAtt](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html).

#### URL

URL to the repository
