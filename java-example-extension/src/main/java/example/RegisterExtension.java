package example;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Optional;

/**
 * Utility class that takes care of registration of extension and fetching the next event
 */
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
public class RegisterExtension {
    private static final String BASEURL = String.format("http://%s/2020-01-01/extension", System.getenv("AWS_LAMBDA_RUNTIME_API"));
    private static final String BODY = "{" +
            "            \"events\": [" +
            "                \"INVOKE\"," +
            "                \"SHUTDOWN\"" +
            "            ]" +
            "        }";

    /**
     * Registers the external extension to listen to "INVOKE" and "SHUTDOWN"
     *
     * @return ID of the registered extension
     */
    public static String registerExtension() {
        final String registerUrl = String.format("%s/register", BASEURL);
        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
                .POST(HttpRequest.BodyPublishers.ofString(BODY))
                .header("Content-Type", "application/json")
                .header("Lambda-Extension-Name", "java-example-extension")
                .uri(URI.create(registerUrl))
                .build();
        try {
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

            // Get extension ID from the response headers
            Optional<String> lambdaExtensionHeader = response.headers().firstValue("lambda-extension-identifier");
            if (lambdaExtensionHeader.isPresent()) return lambdaExtensionHeader.get();
        } catch (Exception e) {
            System.err.println("Error while registering extension: " + e.getMessage());
            e.printStackTrace();
        }

        return null;
    }

    /**
     * Get next event as we have registered for INVOKE event
     *
     * @param extensionId ID of the extension received as response from registration event
     * @return event payload
     */
    public static String getNext(final String extensionId) {
        try {
            final String nextEventUrl = String.format("%s/event/next", BASEURL);
            HttpClient client = HttpClient.newHttpClient();
            HttpRequest request = HttpRequest.newBuilder()
                    .GET()
                    .header("Lambda-Extension-Identifier", extensionId)
                    .uri(URI.create(nextEventUrl))
                    .build();
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() == 200)
                return response.body();
            else
                System.err.printf("%n Invalid status code %s returned while processing event, response %s",
                        response.statusCode(), response.body());
        } catch (Exception e) {
            System.err.println("Error while fetching next event: " + e.getMessage());
            e.printStackTrace();
        }

        return null;
    }
}
