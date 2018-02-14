RETRIES=60

for r in $(seq $RETRIES); do
        if [[ "${KUBERNETES}" ]]; then
                ENDPOINT=$(kubectl get endpoints | grep mediawiki | awk '{ print $2 }')/index.php/Main_Page
        else
                ENDPOINT=$(oc get route | grep mediawiki | cut -f 4 -d ' ')/index.php/Main_Page
        fi

        echo "Running: curl ${ENDPOINT}| grep \"div class\" | cut -f 2 -d \"'\""

        oc get pods
        oc get routes

        BIND_CHECK=$(curl ${ENDPOINT}| grep "div class" | cut -f 2 -d "'")
        if [ "${BIND_CHECK}" = "" ] || [ "${BIND_CHECK}" = "error" ]; then
                echo "Failed to gather data from ${ENDPOINT}"
        else
                echo "SUCCESS"
                echo "You can double check by opening http://${ENDPOINT} in your browser"
                break
        fi
        sleep 2
done

if [ "${r}" == "${RETRIES}" ]; then
        echo "Error: Timeout waiting for verification"
        exit 1
fi
