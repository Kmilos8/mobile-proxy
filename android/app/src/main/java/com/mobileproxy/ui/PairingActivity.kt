package com.mobileproxy.ui

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Bundle
import android.provider.Settings
import android.util.Log
import android.view.View
import android.widget.*
import androidx.activity.result.contract.ActivityResultContracts
import androidx.annotation.OptIn
import androidx.appcompat.app.AppCompatActivity
import androidx.camera.core.*
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.core.content.ContextCompat
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.barcode.common.Barcode
import com.google.mlkit.vision.common.InputImage
import com.mobileproxy.BuildConfig
import com.mobileproxy.R
import com.mobileproxy.core.config.CredentialManager
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import java.io.IOException
import javax.inject.Inject

@AndroidEntryPoint
class PairingActivity : AppCompatActivity() {

    companion object {
        private const val TAG = "PairingActivity"
    }

    @Inject lateinit var credentialManager: CredentialManager

    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var isPairing = false
    private var cameraProvider: ProcessCameraProvider? = null

    private val cameraPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted ->
        if (granted) {
            startCamera()
        } else {
            showManualEntry()
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_pairing)

        // Tab switching
        findViewById<Button>(R.id.btnScanTab).setOnClickListener {
            showScannerView()
        }
        findViewById<Button>(R.id.btnManualTab).setOnClickListener {
            showManualEntry()
        }

        // Manual pair button
        findViewById<Button>(R.id.btnPair).setOnClickListener {
            val serverUrl = findViewById<EditText>(R.id.editPairingServerUrl).text.toString().trim()
            val code = findViewById<EditText>(R.id.editPairingCode).text.toString().trim()
            if (serverUrl.isEmpty() || code.isEmpty()) {
                showError("Please enter both server URL and pairing code")
                return@setOnClickListener
            }
            pair(serverUrl, code)
        }

        // Start with scanner if camera permission is available
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED) {
            startCamera()
        } else {
            cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
        }
    }

    private fun showScannerView() {
        findViewById<View>(R.id.scannerContainer).visibility = View.VISIBLE
        findViewById<View>(R.id.manualContainer).visibility = View.GONE
        findViewById<Button>(R.id.btnScanTab).alpha = 1f
        findViewById<Button>(R.id.btnManualTab).alpha = 0.5f
    }

    private fun showManualEntry() {
        findViewById<View>(R.id.scannerContainer).visibility = View.GONE
        findViewById<View>(R.id.manualContainer).visibility = View.VISIBLE
        findViewById<Button>(R.id.btnScanTab).alpha = 0.5f
        findViewById<Button>(R.id.btnManualTab).alpha = 1f
    }

    @OptIn(ExperimentalGetImage::class)
    private fun startCamera() {
        showScannerView()
        val previewView = findViewById<PreviewView>(R.id.previewView)

        val cameraProviderFuture = ProcessCameraProvider.getInstance(this)
        cameraProviderFuture.addListener({
            cameraProvider = cameraProviderFuture.get()

            val preview = Preview.Builder().build().also {
                it.setSurfaceProvider(previewView.surfaceProvider)
            }

            val barcodeScanner = BarcodeScanning.getClient()

            val imageAnalysis = ImageAnalysis.Builder()
                .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
                .build()

            imageAnalysis.setAnalyzer(ContextCompat.getMainExecutor(this)) { imageProxy ->
                val mediaImage = imageProxy.image ?: run {
                    imageProxy.close()
                    return@setAnalyzer
                }

                val inputImage = InputImage.fromMediaImage(mediaImage, imageProxy.imageInfo.rotationDegrees)

                barcodeScanner.process(inputImage)
                    .addOnSuccessListener { barcodes ->
                        for (barcode in barcodes) {
                            if (barcode.valueType == Barcode.TYPE_TEXT || barcode.valueType == Barcode.TYPE_URL) {
                                val rawValue = barcode.rawValue ?: continue
                                if (rawValue.startsWith("mobileproxy://pair")) {
                                    handleQrCode(rawValue)
                                    break
                                }
                            }
                        }
                    }
                    .addOnCompleteListener {
                        imageProxy.close()
                    }
            }

            try {
                cameraProvider?.unbindAll()
                cameraProvider?.bindToLifecycle(this, CameraSelector.DEFAULT_BACK_CAMERA, preview, imageAnalysis)
            } catch (e: Exception) {
                Log.e(TAG, "Camera bind failed", e)
                showManualEntry()
            }
        }, ContextCompat.getMainExecutor(this))
    }

    private fun handleQrCode(qrValue: String) {
        if (isPairing) return
        Log.i(TAG, "QR code scanned: $qrValue")

        val uri = Uri.parse(qrValue)
        val server = uri.getQueryParameter("server")
        val code = uri.getQueryParameter("code")

        if (server.isNullOrEmpty() || code.isNullOrEmpty()) {
            showError("Invalid QR code format")
            return
        }

        // Stop camera
        cameraProvider?.unbindAll()

        pair(server, code)
    }

    private fun pair(serverUrl: String, code: String) {
        if (isPairing) return
        isPairing = true

        val statusText = findViewById<TextView>(R.id.textPairingStatus)
        val progressBar = findViewById<ProgressBar>(R.id.pairingProgress)
        val pairButton = findViewById<Button>(R.id.btnPair)

        statusText.text = "Pairing..."
        statusText.visibility = View.VISIBLE
        progressBar.visibility = View.VISIBLE
        pairButton.isEnabled = false

        val androidId = Settings.Secure.getString(contentResolver, Settings.Secure.ANDROID_ID)
        val deviceModel = android.os.Build.MODEL
        val androidVersion = android.os.Build.VERSION.RELEASE
        val appVersion = BuildConfig.VERSION_NAME

        val json = JSONObject().apply {
            put("code", code.replace("-", "").uppercase())
            put("android_id", androidId)
            put("device_model", deviceModel)
            put("android_version", androidVersion)
            put("app_version", appVersion)
        }

        val client = OkHttpClient()
        val url = "${serverUrl.trimEnd('/')}/api/public/pair"
        val body = json.toString().toRequestBody("application/json".toMediaType())
        val request = Request.Builder().url(url).post(body).build()

        client.newCall(request).enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                runOnUiThread {
                    isPairing = false
                    progressBar.visibility = View.GONE
                    pairButton.isEnabled = true
                    showError("Connection failed: ${e.message}")
                }
            }

            override fun onResponse(call: Call, response: Response) {
                val responseBody = response.body?.string()
                runOnUiThread {
                    isPairing = false
                    progressBar.visibility = View.GONE
                    pairButton.isEnabled = true

                    if (!response.isSuccessful) {
                        val errorMsg = try {
                            JSONObject(responseBody ?: "").optString("error", "Pairing failed")
                        } catch (_: Exception) { "Pairing failed (${response.code})" }
                        showError(errorMsg)
                        return@runOnUiThread
                    }

                    try {
                        val resp = JSONObject(responseBody ?: "{}")
                        credentialManager.saveCredentials(
                            serverUrl = serverUrl.trimEnd('/'),
                            deviceId = resp.getString("device_id"),
                            authToken = resp.getString("auth_token"),
                            vpnConfig = resp.optString("vpn_config", ""),
                            basePort = resp.optInt("base_port", 0)
                        )

                        statusText.text = "Paired successfully!"
                        setResult(RESULT_OK)
                        finish()
                    } catch (e: Exception) {
                        showError("Invalid response: ${e.message}")
                    }
                }
            }
        })
    }

    private fun showError(message: String) {
        val statusText = findViewById<TextView>(R.id.textPairingStatus)
        statusText.text = message
        statusText.visibility = View.VISIBLE
        statusText.setTextColor(getColor(android.R.color.holo_red_light))
    }

    override fun onDestroy() {
        scope.cancel()
        cameraProvider?.unbindAll()
        super.onDestroy()
    }
}
