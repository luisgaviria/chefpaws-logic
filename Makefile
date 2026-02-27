# chefpaws-logic/Makefile

dev:
	# 1. Start the Drupal backend
	cd ../../chefpaws-drupal && ddev start
	
	# 2. Run the Go Logic Engine
	# We force IS_DDEV_PROJECT=true here so images always work on your MacBook
	export $$(cat .env | xargs) && IS_DDEV_PROJECT=true go run cmd/server/main.go & 
	
	# 3. Wait for Go to boot, then start the Astro frontend
	sleep 3 && cd ../../chefpaws/chefpaws-frontend && npm run dev