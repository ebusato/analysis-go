void summary(TString fileName0, TString fileName1="", TString fileName2="", TString fileName3="", TString fileName4="") {
	gStyle->SetOptStat(0);
	gStyle->SetOptTitle(0);
	
	TChain ch("tree");
	ch.Add(fileName0);
	if(fileName1 != "") {
		ch.Add(fileName1);
	}
	if(fileName2 != "") {
		ch.Add(fileName2);
	}
	if(fileName3 != "") {
		ch.Add(fileName3);
	}
	if(fileName4 != "") {
		ch.Add(fileName4);
	}
	TTreeReader reader(&ch);
        TTreeReaderValue<UInt_t> Run(reader, "Run");
        TTreeReaderValue<UInt_t> Evt(reader, "Evt");
        TTreeReaderArray<UShort_t> IChanAbs240(reader, "IChanAbs240");
        TTreeReaderArray<Double_t> Ampl(reader, "Ampl");
        TTreeReaderArray<Double_t> Charge(reader, "Charge");
	
	TCanvas* c1 = new TCanvas("c1", "c1", 1200, 1200);
	c1->Divide(3,3);
	c1->cd(1);
	ch.Draw("IChanAbs240");
	c1->cd(2);
	ch.Draw("IQuartetAbs60");
	c1->cd(3);
	ch.Draw("IQuartetAbs60[1] : IQuartetAbs60[0]", "", "colz");
	c1->cd(4);
	ch.Draw("E");
	c1->cd(5);
	ch.Draw("T30[0] - T30[1]>>hCRT(200,-10,10)");
	TH1F* hCRT = (TH1F*) gDirectory->Get("hCRT");
	hCRT->Fit("gaus");
	hCRT->SetStats(1);
}