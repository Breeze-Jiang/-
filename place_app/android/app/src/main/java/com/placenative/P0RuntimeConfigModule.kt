package com.placenative

import com.facebook.react.bridge.ReactApplicationContext
import com.facebook.react.bridge.ReactContextBaseJavaModule
import com.facebook.react.bridge.NativeModule
import com.facebook.react.ReactPackage
import com.facebook.react.uimanager.ViewManager

class P0RuntimeConfigModule(reactContext: ReactApplicationContext) : ReactContextBaseJavaModule(reactContext) {
  override fun getName(): String = "P0RuntimeConfig"

  override fun getConstants(): Map<String, Any> = mapOf("apiOrigin" to BuildConfig.P0_API_ORIGIN)
}

class P0RuntimeConfigPackage : ReactPackage {
  override fun createNativeModules(reactContext: ReactApplicationContext): List<NativeModule> =
      listOf(P0RuntimeConfigModule(reactContext))

  override fun createViewManagers(reactContext: ReactApplicationContext): List<ViewManager<*, *>> = emptyList()
}
