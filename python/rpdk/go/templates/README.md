# {{ type_name }}

Congratulations on starting development! Next steps:

1. Write the JSON schema describing your resource, `{{ schema_path.name }}`
2. The RPDK will automatically generate the correct resource model from the
   schema whenever the project is built via Maven. You can also do this manually
   with the following command: `{{ executable }} generate`
3. Implement your resource handlers by adding code to provision your resources in the various Handler classes.
4. Deployment is best achieved with the .NET Core CLI;
```
dotnet tool install -g Amazon.Lambda.Tools
dotnet build -c Release -r rhel.7.2-x64
dotnet lambda package --framework netcoreapp2.0
```

Please don't modify files under `{{ generated_root }}`, as they will be
automatically overwritten.