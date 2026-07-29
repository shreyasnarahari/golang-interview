# Topic 13 Medium Questions

### Q: How do you read struct field tags using reflection?
**Answer**: Using `reflect.TypeOf(s).Field(i).Tag.Get("json")`.
