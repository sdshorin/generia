# gRPC Generated Code

This directory contains pre-generated gRPC Python code for the Generia platform microservices communication.

## Files Structure

- `character/` - Character service gRPC stubs
- `media/` - Media service gRPC stubs
- `post/` - Post service gRPC stubs
- `world/` - World service gRPC stubs

## How These Files Were Generated

These files were generated using the `generate_local_grpc.sh` script in the `scripts` directory. The script:

1. Takes protobuf files from `/api/proto/`
2. Generates Python code using the `protoc` compiler
3. Fixes import statements to work correctly with Python's module system
4. Creates necessary `__init__.py` files

## Regenerating Files

If proto files change, you can regenerate the gRPC code by running:

```bash
# From the root of the repo
./services/ai-worker/scripts/generate_local_grpc.sh
```

Then commit the changed files to the repository.

## Why Pre-generate Instead of Generating at Runtime?

We pre-generate gRPC files for several reasons:

1. **Reliability** - Eliminates potential errors during container startup
2. **Speed** - Faster container startup time
3. **Version Control** - Changes to gRPC interfaces are tracked in git
4. **Simplicity** - No need for complex runtime code generation logic
5. **Testing** - Can verify the generated code works before deployment