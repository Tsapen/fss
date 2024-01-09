export FSS_ROOT_DIR = $(shell pwd)
export FSS_CONFIG = /configs/fss_config.json
export FSS_MIGRATIONS_PATH = /migrations/
export FSS_CLIENT_CONFIG = /configs/client_config.json
export PORT = $(shell grep -o '"address": "[^"]*"' $(FSS_ROOT_DIR)$(FSS_CONFIG) | cut -d ':' -f 3 | cut -d '"' -f 1)

export FSS_TEST_MIGRATIONS_PATH = /migrations/test/
export FSS_TEST_CONFIG = /configs/test_fss_config.json
export FS_TEST_CONFIG = /configs/test_fs_config.json
export FSS_TEST_CLIENT_CONFIG = /configs/test_fss_config.json

test:
	$(call print-target)
	docker-compose -f $(FSS_ROOT_DIR)/deployment/test/docker-compose.yml down --remove-orphans
	docker-compose -f $(FSS_ROOT_DIR)/deployment/test/docker-compose.yml up -d --build test-fss

run:
	rm -rf $(FSS_ROOT_DIR)/socket
	docker-compose -f $(FSS_ROOT_DIR)/deployment/fss/docker-compose.yml down --remove-orphans
	docker-compose -f $(FSS_ROOT_DIR)/deployment/fss/docker-compose.yml up --build fss --detach

stop:
	docker-compose -f $(FSS_ROOT_DIR)/deployment/fss/docker-compose.yml down -v --remove-orphans
