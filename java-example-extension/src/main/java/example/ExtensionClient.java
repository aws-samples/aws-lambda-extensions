// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package example;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Optional;

/**
 * Utility class that takes care of registration of extension, fetching the next event, initializing
 * and exiting with error
 */
public class ExtensionClient {
	private static final String EXTENSION_NAME = "java-example-extension";
	private static final String BASEURL = String
			.format("http://%s/2020-01-01/extension", System.getenv("AWS_LAMBDA_RUNTIME_API"));
	private static final String BODY = "{" +
			"            \"events\": [" +
			"                \"INVOKE\"," +
			"                \"SHUTDOWN\"" +
			"            ]" +
			"        }";
	private static final String LAMBDA_EXTENSION_IDENTIFIER = "Lambda-Extension-Identifier";
	private static final String LAMBDA_EXTENSION_FUNCTION_ERROR_TYPE = "Lambda-Extension-Function-Error-Type";
	private static final HttpClient client = HttpClient.newHttpClient();

	/**
	 * Registers the external extension to listen to "INVOKE" and "SHUTDOWN"
	 *
	 * @return ID of the registered extension
	 */
	public static String registerExtension() {
		final String registerUrl = String.format("%s/register", BASEURL);
		HttpRequest request = HttpRequest.newBuilder()
				.POST(HttpRequest.BodyPublishers.ofString(BODY))
				.header("Content-Type", "application/json")
				.header("Lambda-Extension-Name", EXTENSION_NAME)
				.uri(URI.create(registerUrl))
				.build();
		try {
			HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

			// Get extension ID from the response headers
			Optional<String> lambdaExtensionHeader = response.headers().firstValue("lambda-extension-identifier");
			if (lambdaExtensionHeader.isPresent()) return lambdaExtensionHeader.get();
		}
		catch (Exception e) {
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
			HttpRequest request = HttpRequest.newBuilder()
					.GET()
					.header(LAMBDA_EXTENSION_IDENTIFIER, extensionId)
					.uri(URI.create(nextEventUrl))
					.build();
			HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
			if (response.statusCode() == 200)
				return response.body();
			else
				System.err.printf("%n Invalid status code %s returned while processing event, response %s",
						response.statusCode(), response.body());
		}
		catch (Exception e) {
			System.err.println("Error while fetching next event: " + e.getMessage());
			e.printStackTrace();
		}

		return null;
	}

	/**
	 * InitError reports an initialization error to the platform. Call it when you registered but failed to initialize
	 * @param extensionId ID of the extension received as response from registration event
	 * @param errorType error type
	 * @return response body
	 */
	public static String initError(final String extensionId, final String errorType) {
		try {
			final String nextEventUrl = String.format("%s/init/error", BASEURL);
			HttpRequest request = HttpRequest.newBuilder()
					.POST(null)
					.header(LAMBDA_EXTENSION_IDENTIFIER, extensionId)
					.header(LAMBDA_EXTENSION_FUNCTION_ERROR_TYPE, errorType)
					.uri(URI.create(nextEventUrl))
					.build();
			HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
			if (response.statusCode() == 200)
				return response.body();
			else
				System.err.printf("%n Invalid status code %s returned will processing the request, response %s",
						response.statusCode(), response.body());
		}
		catch (Exception e) {
			System.err.println("Error while initializing error event: " + e.getMessage());
			e.printStackTrace();
		}

		return null;
	}

	/**
	 * ExitError reports an error to the platform before exiting. Call it when you encounter an unexpected failure
	 * @param extensionId ID of the extension received as response from registration event
	 * @param errorType error type
	 * @return response body
	 */
	public static String exitError(final String extensionId, final String errorType) {
		try {
			final String nextEventUrl = String.format("%s/exit/error", BASEURL);
			HttpRequest request = HttpRequest.newBuilder()
					.POST(null)
					.header(LAMBDA_EXTENSION_IDENTIFIER, extensionId)
					.header(LAMBDA_EXTENSION_FUNCTION_ERROR_TYPE, errorType)
					.uri(URI.create(nextEventUrl))
					.build();
			HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
			if (response.statusCode() == 200)
				return response.body();
			else
				System.err.printf("%n Invalid status code %s returned will processing the request, response %s",
						response.statusCode(), response.body());
		}
		catch (Exception e) {
			System.err.println("Error while exiting with error event: " + e.getMessage());
			e.printStackTrace();
		}

		return null;
	}
}
