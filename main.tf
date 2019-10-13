provider "dockerbuild" {
}

data "dockerbuild_source_tree" "sample_service" {
    source_dirs = ["sample_service"]
}

resource "dockerbuild_build" "sample_service" {
    source_dir = "${path.module}/sample_service"
    source_hash = data.dockerbuild_source_tree.sample_service.hash
}