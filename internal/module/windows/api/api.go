// +build windows

package api

import (
	"golang.org/x/sys/windows"
)

var (
	modNTDLL    = windows.NewLazySystemDLL("ntdll.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	modIphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")
	modBcrypt   = windows.NewLazySystemDLL("bcrypt.dll")

	procNTQueryInformationProcess = modNTDLL.NewProc("NtQueryInformationProcess")

	procReadProcessMemory   = modKernel32.NewProc("ReadProcessMemory")
	procWriteProcessMemory  = modKernel32.NewProc("WriteProcessMemory")
	procVirtualAllocEx      = modKernel32.NewProc("VirtualAllocEx")
	procVirtualFreeEx       = modKernel32.NewProc("VirtualFreeEx")
	procVirtualProtectEx    = modKernel32.NewProc("VirtualProtectEx")
	procCreateRemoteThread  = modKernel32.NewProc("CreateRemoteThread")
	procGetSystemInfo       = modKernel32.NewProc("GetSystemInfo")
	procGetNativeSystemInfo = modKernel32.NewProc("GetNativeSystemInfo")

	procGetExtendedTCPTable = modIphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUDPTable = modIphlpapi.NewProc("GetExtendedUdpTable")

	procBCryptOpenAlgorithmProvider  = modBcrypt.NewProc("BCryptOpenAlgorithmProvider")
	procBCryptCloseAlgorithmProvider = modBcrypt.NewProc("BCryptCloseAlgorithmProvider")
	procBCryptSetProperty            = modBcrypt.NewProc("BCryptSetProperty")
	procBCryptGetProperty            = modBcrypt.NewProc("BCryptGetProperty")
	procBCryptGenerateSymmetricKey   = modBcrypt.NewProc("BCryptGenerateSymmetricKey")
	procBCryptDestroyKey             = modBcrypt.NewProc("BCryptDestroyKey")
	procBCryptDecrypt                = modBcrypt.NewProc("BCryptDecrypt")
)

// CloseHandle is used to close handle it will return error.
func CloseHandle(handle windows.Handle) {
	_ = windows.CloseHandle(handle)
}