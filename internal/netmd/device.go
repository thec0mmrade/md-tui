package netmd

import "github.com/google/gousb"

type Device struct {
	vendorId gousb.ID
	deviceId gousb.ID
	Name     string
}

var (
	Devices = [...]Device{
		{vendorId: 0x04dd, deviceId: 0x7202, Name: "Sharp IM-MT899H"},
		{vendorId: 0x04dd, deviceId: 0x9013, Name: "Sharp IM-DR400/DR410/DR420"},
		{vendorId: 0x04dd, deviceId: 0x9014, Name: "Sharp IM-DR80"},
		{vendorId: 0x054c, deviceId: 0x0034, Name: "Sony PCLK-XX"},
		{vendorId: 0x054c, deviceId: 0x0036, Name: "Sony"},
		{vendorId: 0x054c, deviceId: 0x0075, Name: "Sony MZ-N1"},
		{vendorId: 0x054c, deviceId: 0x007c, Name: "Sony"},
		{vendorId: 0x054c, deviceId: 0x0080, Name: "Sony LAM-1"},
		{vendorId: 0x054c, deviceId: 0x0081, Name: "Sony MDS-JB980/JE780"},
		{vendorId: 0x054c, deviceId: 0x0084, Name: "Sony MZ-N505"},
		{vendorId: 0x054c, deviceId: 0x0085, Name: "Sony MZ-S1"},
		{vendorId: 0x054c, deviceId: 0x0086, Name: "Sony MZ-N707"},
		{vendorId: 0x054c, deviceId: 0x008e, Name: "Sony CMT-C7NT"},
		{vendorId: 0x054c, deviceId: 0x0097, Name: "Sony PCGA-MDN1"},
		{vendorId: 0x054c, deviceId: 0x00ad, Name: "Sony CMT-L7HD"},
		{vendorId: 0x054c, deviceId: 0x00c6, Name: "Sony MZ-N10"},
		{vendorId: 0x054c, deviceId: 0x00c7, Name: "Sony MZ-N910"},
		{vendorId: 0x054c, deviceId: 0x00c8, Name: "Sony MZ-N710/NF810"},
		{vendorId: 0x054c, deviceId: 0x00c9, Name: "Sony MZ-N510/N610"},
		{vendorId: 0x054c, deviceId: 0x00ca, Name: "Sony MZ-NE410/NF520D"},
		{vendorId: 0x054c, deviceId: 0x00eb, Name: "Sony MZ-NE810/NE910"},
		{vendorId: 0x054c, deviceId: 0x0101, Name: "Sony LAM-10"},
		{vendorId: 0x054c, deviceId: 0x0113, Name: "Aiwa AM-NX1"},
		{vendorId: 0x054c, deviceId: 0x013f, Name: "Sony MDS-S500"},
		{vendorId: 0x054c, deviceId: 0x014c, Name: "Aiwa AM-NX9"},
		{vendorId: 0x054c, deviceId: 0x017e, Name: "Sony MZ-NH1"},
		{vendorId: 0x054c, deviceId: 0x0180, Name: "Sony MZ-NH3D"},
		{vendorId: 0x054c, deviceId: 0x0182, Name: "Sony MZ-NH900"},
		{vendorId: 0x054c, deviceId: 0x0184, Name: "Sony MZ-NH700/NH800"},
		{vendorId: 0x054c, deviceId: 0x0186, Name: "Sony MZ-NH600"},
		{vendorId: 0x054c, deviceId: 0x0187, Name: "Sony MZ-NH600D"},
		{vendorId: 0x054c, deviceId: 0x0188, Name: "Sony MZ-N920"},
		{vendorId: 0x054c, deviceId: 0x018a, Name: "Sony LAM-3"},
		{vendorId: 0x054c, deviceId: 0x01e9, Name: "Sony MZ-DH10P"},
		{vendorId: 0x054c, deviceId: 0x0219, Name: "Sony MZ-RH10"},
		{vendorId: 0x054c, deviceId: 0x021b, Name: "Sony MZ-RH710/MZ-RH910"},
		{vendorId: 0x054c, deviceId: 0x021d, Name: "Sony CMT-AH10"},
		{vendorId: 0x054c, deviceId: 0x022c, Name: "Sony CMT-AH10"},
		{vendorId: 0x054c, deviceId: 0x023c, Name: "Sony DS-HMD1"},
		{vendorId: 0x054c, deviceId: 0x0286, Name: "Sony MZ-RH1"},
	}
)
